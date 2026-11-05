# How to export for slides

A diagram in a deck wants the drawing only, without the header, the cards and the notes.

```bash
noodle render architecture/platform -view runtime -theme light -slide -o runtime.slide.drawio
drawio -x -f png -s 2 --border 20 -o runtime.slide.png runtime.slide.drawio
```

`-slide` crops the canvas to the drawing and leaves out the title, the subtitle, the cards
and the notes. Choose the theme that matches the slide background.

For documentation that people may want to edit later, export an SVG that embeds the
diagram:

```bash
drawio -x -f svg --embed-diagram -o runtime.drawio.svg runtime.light.drawio
```

Mind the sizes: a PNG exported at `-s 2` of a large topology is several megabytes, and an
embedded SVG is larger still, because draw.io embeds a PNG fallback for every label. Use
the PNG in chat and slides, and the SVG where editing matters.
