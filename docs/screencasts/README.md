# Screencast / Asciinema

[Asciinema](https://github.com/asciinema/asciinema) is used to automatically generate
terminal session recordings for our documentation. To fully automate this we use scripts
that utilize [expect](https://linux.die.net/man/1/expect) to interface with different
CLI tools, and run them inside a [container](docker/Dockerfile).

## Usage

```sh
./generate-screencasts.sh
```

This will:
+ build the container
+ run the expect based scripts
+ copy recordings into the assets folder of our docs

To replay the output you can use `asciinema play recordings/verify-cli.cast`.

Include the generated screencast into our docs using the [`AsciinemaWidget`](../src/components/AsciinemaWidget/index.js):

```md
import AsciinemaWidget from '../../src/components/AsciinemaWidget';

<AsciinemaWidget src="/constellation/assets/verify-cli.cast" fontSize={16} rows={18} cols={80} idleTimeLimit={3} preload={true} theme={'edgeless'} />
```

Then [re-build and locally host the docs](../README.md).

## Styling

There are three different locations were styling is applied:

1. **The prompt** is styled using [ANSI escape codes](https://en.wikipedia.org/wiki/ANSI_escape_code).
More explanation and the actual color codes can be found in [Dockerfile](docker/Dockerfile).
2. **Player dimensions** are passed to the [`AsciinemaWidget`](../src/components/AsciinemaWidget/index.js)
when it is [embedded in the docs](../docs/workflows/verify-cli.md#5). Check the `asciinema-player` for a
[full list of options](https://github.com/asciinema/asciinema-player#options).
3. **Everything else** is [styled via CSS](../src/css/custom.css). This includes the option to build a custom
[player theme](https://github.com/asciinema/asciinema-player/wiki/Custom-terminal-themes).

###

## GitHub README.md

The GitHub `README.md` does not support embedding the `asciinema-player`, therefore we generate an
`svg` file for that usecase.

{"version": 2, "width": 0, "height": 0, "timestamp": 1676289328, "env": {"SHELL": "/bin/bash", "TERM": "xterm-256color"}}

=> TODO: Automate that change

{"version": 2, "width": 95, "height": 17, "timestamp": 1676289328, "env": {"SHELL": "/bin/bash", "TERM": "xterm-256color"}}

```sh
# https://github.com/nbedos/termtosvg
# Archived since 2020, do we want to change?
pip3 install termtosvg
# Window-frame.svg contains the styling information
termtosvg render recordings/readme.cast readme.svg -t window-frame.svg
cp readme.svg ../static/img/shell-windowframe.svg
```
