# pngarbage

A CLI utility to scan web pages for garbage PNGs.

### What?

Improper image format usage can be a serious concern for frontend web performance. PNGs can be particularly problematic as they are typically much larger than JPEGs.

pngarbage will scan a URL and attempt to identify cases where PNGs are used when they really shouldn't be

### How?

Currently, pngarbage will flag a png as "garbage" if it has no transparent pixels. Alpha transparency is one of the primary reasons to use a PNG and if an image doesn't have any transparent pixels, there's a high likelihood that JPEG could have been used, resulting in serious savings.

### Usage

Download the latest binary from the [releases page](https://github.com/mpchadwick/pngarbage/releases). Then, simply specify the URL to scan and run it

```
➜  ./pngarbage -url="http://localhost:8080"
===========================
> pngarbage
===========================
Checking:  http://localhost:8080
Number of pngs:  2
http://localhost:8080/sample.png  is garbage! Content-Length:  31520
```
