package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"
)

// ffmpeg -nostdin -i "$name" -c:a alac -c:v copy "${name%.*}.m4a"

var (
	baseDir = flag.String("d", "", "The base directory containing the FLAC files")
)

var (
	// ffmpeg -loglevel error -nostdin -i $IN_FILE -c:a alac -c:v copy $OUT_FILE
	args = []string{
		"-loglevel", "error",
		"-nostdin",
		"-i", "",
		"-c:a", "alac",
		"-c:v", "copy",
		"",
	}
)

func getCount(ctx context.Context) (int, error) {
	count := 0
	err := filepath.WalkDir(*baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".flac") {
			count++
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to get count: %v", err)
	}
	return count, nil
}

func convert(ctx context.Context, total int) error {
	g, ctx := errgroup.WithContext(ctx)
	paths := make(chan string)

	g.Go(func() error {
		defer close(paths)
		defer fmt.Println()
		i := 0
		return filepath.WalkDir(*baseDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".flac") {
				fmt.Printf("\rProgress %d/%d (%0.2f)", i, total, float64(i)/float64(total))
				i++
				select {
				case <-ctx.Done():
					return ctx.Err()
				case paths <- path:
				}
			}
			return nil
		})
	})

	const numConv = 20
	for i := 0; i < numConv; i++ {
		g.Go(func() error {
			for path := range paths {
				cmdArgs := make([]string, len(args))
				copy(cmdArgs, args)
				dir, name := filepath.Split(path)
				cmdArgs[4] = path
				cmdArgs[9] = dir + name[:len(name)-5] + ".m4a"
				cmd := exec.Command("ffmpeg", cmdArgs...)
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("failed to convert %q: %v", path, err)
				}
				if err := os.Remove(path); err != nil {
					fmt.Printf("failed to remove %q: %v\n", path, err)
				}
			}
			return nil
		})
	}
	return g.Wait()
}

func main() {
	if *baseDir == "" {
		log.Fatal("Missing required '-d' flag")
	}
	ctx := context.Background()
	c, err := getCount(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if err := convert(ctx, c); err != nil {
		log.Fatal(err)
	}
}
