# img-fwd Demo

Live demo showcasing img-fwd image optimization capabilities.

## Live URL

**Demo site:** https://img-fwd.driedel.dev  
**Proxy endpoint:** https://cdn.img-fwd.driedel.dev

## What it shows

- **Comparative view**: Original images vs optimized (AVIF/WebP) with size savings
- **Before/After slider**: Interactive comparison of quality
- **Transformation gallery**: Resize, format conversion, quality control, blur effects
- **Real-time stats**: File sizes measured via Resource Timing API

## AI Image Prompts

Generate these images in high resolution (min 1920px wide) for best results:

| File | Format | Prompt |
|---|---|---|
| `photo.jpg` | JPEG | *"A stunning aerial photograph of a coastal landscape at golden hour, vivid colors, turquoise ocean, rocky cliffs, professional photography, ultra high resolution, sharp details"* |
| `ui-mockup.png` | PNG | *"Clean modern UI dashboard mockup with charts and data cards, flat design, white background, high resolution screenshot, crisp typography, professional SaaS interface"* |
| `icon.svg` | SVG | *"Minimalist geometric logo icon, abstract shape, single color, vector style, clean lines, suitable for favicon or app icon"* |
| `animated.gif` | GIF | *"A smooth loading animation loop, pulsing geometric shapes morphing into each other, seamless loop, 10 frames, minimalist style, pastel colors"* |
| `banner.webp` | WebP/PNG | *"Wide panoramic banner image of a futuristic city skyline at dusk, neon lights, reflections on water, cinematic composition, ultra wide aspect ratio 21:9"* |
| `portrait.jpg` | JPEG | *"Professional portrait photography of a person in natural light, shallow depth of field, bokeh background, high detail skin texture, studio quality"* |

> **Note**: If AI doesn't generate exact formats (especially SVG and animated GIF), generate as PNG/MP4 and convert locally.

## Local testing

Open `index.html` directly in a browser (uses relative paths for CSS/JS).  
Images loaded from `cdn.img-fwd.driedel.dev` require the proxy to be running.

## Architecture

```
Browser → img-fwd.driedel.dev (GitHub Pages) → serves demo site
Browser → cdn.img-fwd.driedel.dev (Fly.io) → img-fwd proxy → imgproxy
                                         ↓
                              fetches originals from
                              img-fwd.driedel.dev (GitHub Pages)
```

## Deploy checklist

- [ ] Generate and add images to `demo/images/`
- [ ] Enable GitHub Pages (source: `/demo` folder, custom domain: `img-fwd.driedel.dev`)
- [ ] Create Fly.io account and deploy: `fly deploy`
- [ ] Configure Cloudflare DNS (see main `AGENTS.md`)
- [ ] Test proxy endpoints returning transformed images
