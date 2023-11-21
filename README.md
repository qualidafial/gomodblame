# gomodblame

Generate Mermaid graph to visualize how go mod dependencies are
being brought in.

## Installation

```shell
go install github.com/qualidafial/gomodblame@latest
```

## Usage

```shell
gomodblame google.golang.org/grpc@v1.55.0 > graph.mermaid
```

## Example Output

```mermaid
graph LR;
    n0["google.golang.org/grpc@v1.55.0"];
    n1["mycompany.com/my-app"];
    n1 --> n0;
    n2["cloud.google.com/go/compute@v1.19.3"];
    n2 --> n0;
    n1 --> n2;
    n3["google.golang.org/api@v0.125.0"];
    n3 --> n2;
    n1 --> n3;
    n4["mycompany.com/foo@v1.2.3"];
    n4 --> n3;
    n1 --> n4;
    n5["mycompany.com/bar@v4.5.6"];
    n5 --> n4;
    n1 --> n5;
    n5 --> n3;
    n4 --> n2;
    n5 --> n2;
    n3 --> n0;
    n4 --> n0;
    n5 --> n0;
```