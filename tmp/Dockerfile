FROM --platform=amd64 ubuntu:22.04 AS generate
RUN useradd pptruser
RUN apt-get update && apt-get install ffmpeg npm libnss3 --yes
RUN npm install svgexport -g --unsafe-perm=true
COPY frames /tmp/frames
RUN chown pptruser /tmp/frames
USER pptruser
RUN /bin/bash -c 'for file in /tmp/frames/*.svg; do svgexport $file $file.png 3x; done'
RUN ffmpeg -i /tmp/frames/termtosvg_00%03d.svg.png -lossless 1 -r 25 /tmp/animated.webp

FROM scratch AS export
COPY --from=generate /tmp/animated.webp .
