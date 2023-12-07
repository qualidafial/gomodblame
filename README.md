# gomodblame

Generate Mermaid graph to visualize how go mod dependencies are
being brought in.

## Installation

```shell
go install github.com/qualidafial/gomodblame@latest
```

## Usage

```shell
gomodblame toolchain > graph.mermaid
```

## Example Output

```
graph LR;
    n0["github.com/qualidafial/gomodblame"];
    n1["golang.org/x/exp@v0.0.0-20231206192017-f3f8817b8deb"];
    n0 --> n1;
    n2["github.com/google/go-cmp@v0.5.8"];
    n1 --> n2;
    n3["golang.org/x/mod@v0.14.0"];
    n1 --> n3;
    n4["golang.org/x/tools@v0.16.0"];
    n1 --> n4;
```

```mermaid
graph LR;
    n0["github.com/qualidafial/gomodblame"];
    n1["golang.org/x/exp@v0.0.0-20231206192017-f3f8817b8deb"];
    n0 --> n1;
    n2["github.com/google/go-cmp@v0.5.8"];
    n1 --> n2;
    n3["golang.org/x/mod@v0.14.0"];
    n1 --> n3;
    n4["golang.org/x/tools@v0.16.0"];
    n1 --> n4;
```