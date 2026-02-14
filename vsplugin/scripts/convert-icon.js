const sharp = require('sharp');
const fs = require('fs');
const path = require('path');

const svgPath = path.join(__dirname, '..', 'resources', 'icon.svg');
const pngPath = path.join(__dirname, '..', 'resources', 'icon.png');

const svgBuffer = fs.readFileSync(svgPath);

sharp(svgBuffer)
  .resize(256, 256)
  .png()
  .toFile(pngPath)
  .then(() => {
    console.log('Icon converted successfully:', pngPath);
  })
  .catch(err => {
    console.error('Error converting icon:', err);
    process.exit(1);
  });
