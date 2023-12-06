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

```mermaid
graph LR;
    n0["github.com/qualidafial/gomodblame"];
    n1["go@1.21.1"];
    n0 --> n1;
    n2["toolchain@go1.21.1"];
    n1 --> n2;
```