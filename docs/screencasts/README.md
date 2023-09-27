# Screencast / Asciinema

[Asciinema](https://github.com/asciinema/asciinema) is used to automatically generate
terminal session recordings for our documentation. To fully automate this we use scripts
that utilize [expect](https://manpages.debian.org/testing/expect/expect.1.en.html) to interface with different
CLI tools, and run them inside a [container](docker/Dockerfile).

## Usage

```sh
mkdir -p constellation
./generate-screencasts.sh
sudo chown -R $USER:$USER ./constellation
cd constellation && constellation iam destroy
cd .. && rm -rf ./constellation
```

This will:
+ build the container
+ run the expect based scripts
+ copy recordings into the assets folder of our docs

To replay the output you can use `asciinema play recordings/verify-cli.cast`.

Include the generated screencast into our docs using the [`AsciinemaWidget`](../src/components/AsciinemaWidget/index.js):

```md
import AsciinemaWidget from '../../src/components/AsciinemaWidget';

<AsciinemaWidget src="/constellation/assets/verify-cli.cast" fontSize={16} rows={20} cols={112} idleTimeLimit={3} preload={true} theme={'edgeless'} />
```

Then [re-build and locally host the docs](../README.md).

## Styling

There are three different locations were styling is applied:

1. **The prompt** is styled using [ANSI escape codes](https://en.wikipedia.org/wiki/ANSI_escape_code).
More explanation and the actual color codes can be found in [Dockerfile](docker/Dockerfile).
2. **Player dimensions** are passed to the [`AsciinemaWidget`](../src/components/AsciinemaWidget/index.js)
when it's [embedded in the docs](../docs/workflows/verify-cli.md#5). Check the `asciinema-player` for a
[full list of options](https://github.com/asciinema/asciinema-player#options).
1. **Everything else** is [styled via CSS](../src/css/custom.css). This includes the option to build a custom
[player theme](https://github.com/asciinema/asciinema-player/wiki/Custom-terminal-themes).

###

## GitHub README.md

The GitHub `README.md` doesn't support embedding the JavaScript `asciinema-player`, therefore we generate an
`svg` file for that use case.

```sh
# Make sure to install the converter.
# https://github.com/nbedos/termtosvg
pip3 install termtosvg

# Generate SVG. This takes ~10min, since it actually creates a cluster in GCP.
./generate-readme-svg.sh
```
