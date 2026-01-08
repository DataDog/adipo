# Go Hello World Fat Binary Example

This example demonstrates building a Go program for multiple x86-64 micro-architecture levels and packaging them into a fat binary using adipo and Bazel.

## What It Does

The example:
1. Builds the same Go application three times for different x86-64 levels:
   - `x86-64-v1`: Baseline (compatible with all x86-64 CPUs)
   - `x86-64-v2`: Requires SSE4.2, POPCNT (2009+ CPUs)
   - `x86-64-v3`: Requires AVX2, BMI2 (2013+ CPUs)
2. Packages all three binaries into a single fat binary using adipo
3. At runtime, automatically selects the best version for the CPU

## Building

```bash
# Build individual binaries
bazel build //examples/go_hello:hello_v1
bazel build //examples/go_hello:hello_v2
bazel build //examples/go_hello:hello_v3

# Build the fat binary (builds all versions automatically)
bazel build //examples/go_hello:hello_fat

# Run the fat binary
bazel run //examples/go_hello:hello_fat
```

## Inspecting

```bash
# Build adipo first
bazel build //cmd/adipo

# Inspect the fat binary
bazel-bin/cmd/adipo/adipo_/adipo inspect bazel-bin/examples/go_hello/hello_fat
```

## Go and Micro-Architecture Levels

Go 1.18+ supports the `GOAMD64` environment variable to control x86-64 optimization levels:
- `GOAMD64=v1` (default): Baseline x86-64
- `GOAMD64=v2`: Adds SSE4.2, POPCNT
- `GOAMD64=v3`: Adds AVX2, BMI2, etc.
- `GOAMD64=v4`: Adds AVX-512

In production, you would typically:
```bash
# Build with different GOAMD64 values
GOAMD64=v1 go build -o myapp-v1 .
GOAMD64=v2 go build -o myapp-v2 .
GOAMD64=v3 go build -o myapp-v3 .

# Create fat binary
adipo create -o myapp.fat \
  --binary myapp-v1:x86-64-v1 \
  --binary myapp-v2:x86-64-v2 \
  --binary myapp-v3:x86-64-v3
```

## Limitations

Note: The Bazel `rules_go` doesn't currently expose a direct way to set `GOAMD64`.
In this example, all three binaries will be functionally identical. To properly
demonstrate different optimization levels, you would need to:

1. Use a custom toolchain that sets `GOAMD64`
2. Or build with the Go compiler directly and import into Bazel
3. Or use different build tags to conditionally include optimized code

The BUILD.bazel file shows the structure you would use, even though the
optimization levels aren't fully implemented in this example.

## Real-World Use Case

In production, you might use this for:
```python
adipo_fat_binary(
    name = "myservice",
    binaries = {
        "//cmd/myservice:binary_v1": "x86-64-v1",  # For old AWS instances
        "//cmd/myservice:binary_v2": "x86-64-v2",  # For c5 instances
        "//cmd/myservice:binary_v3": "x86-64-v3",  # For c7a instances
    },
)
```

Then deploy the same `myservice` fat binary to all your servers, and it will
automatically run the optimal version on each instance type.
