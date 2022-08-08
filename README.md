# flac2alac

Quick script I wrote to convert a bunch of FLAC files to ALAC.  All it does
is make concurrent calls out to `ffmpeg` which makes it faster than using a
shell loop.

## Usage

Install `ffmpeg`:

```sh
sudo apt install -y ffmpeg
```

Clone this repo and build the binary:

```sh
git clone https://github.com/jessesomerville/flac2alac
cd flac2alac
go mod tidy && go build
```

Run `flac2alac` and provide the base directory containing the FLAC files:

```sh
flac2alac -d '/home/me/Music'
```