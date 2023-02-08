import BrowserOnly from '@docusaurus/BrowserOnly';
import React, { useEffect, useRef } from 'react';
import 'asciinema-player/dist/bundle/asciinema-player.css';

const AsciinemaWidget = ({ src, ...asciinemaOptions}) => {
  return (
    <BrowserOnly fallback={<div>Loading asciinema cast...</div>}>
      {() => {
        const AsciinemaPlayer = require('asciinema-player');
        const ref = useRef(null);

        useEffect(() => {
          AsciinemaPlayer.create(src, ref.current, asciinemaOptions);
        }, [src]);

        return <div ref={ref} />;
      }}
    </BrowserOnly>
  );
};

export default AsciinemaWidget;
