# Compression Optimization Analysis

## Current Approach

Currently, `adipo` compresses each binary independently using zstd level 3:
- Each binary is compressed separately
- No shared context or dictionary between binaries
- Simple and straightforward decompression (extract only what you need)

## Analysis of Binary Similarity

Testing with 4 x86-64 variants (v1, v2, v3, v4) of a simple Go binary:

### Binary Characteristics
```
Original sizes:
  v1: 2.15 MB (x86-64-v1)
  v2: 2.15 MB (x86-64-v2)
  v3: 2.15 MB (x86-64-v3)
  v4: 2.15 MB (x86-64-v4)
  Total: 8.60 MB
```

### Byte-by-byte Similarity
```
v1 vs v2: 13.1% identical bytes
v2 vs v3: 22.7% identical bytes
v3 vs v4: 39.9% identical bytes
```

Despite being compiled from the same source, the binaries have:
- **Low byte-level similarity** (13-40%)
- **No identical 4KB chunks** between versions
- Different instruction encodings for different microarchitecture levels
- Different alignment, padding, and relocation tables

## Compression Results

### 1. Current Approach (Independent Compression)
```
Level 3, each binary compressed independently:
  Total: 5.20 MB compressed (60.4% ratio)

Breakdown:
  v1: 1.36 MB (60.4%)
  v2: 1.36 MB (60.5%)
  v3: 1.36 MB (60.4%)
  v4: 1.36 MB (60.4%)
```

**Pros:**
- Simple implementation (current approach)
- Fast decompression
- Can extract individual binaries without decompressing all
- Good for runtime selection (only decompress what you need)

**Cons:**
- Doesn't leverage similarity between binaries
- Each binary compressed in isolation

### 2. Combined Stream Compression
```
All 4 binaries concatenated and compressed as single stream:
  Total: 3.53 MB compressed (41.0% ratio)

Improvement: 1.75 MB saved (32.2% better!)
```

**Pros:**
- **Massive improvement**: 32.2% better compression
- Zstd can find and exploit redundancy across all binaries
- Best compression ratio possible with zstd

**Cons:**
- Must decompress ALL binaries to extract one
- Higher memory usage during decompression
- Slower runtime selection (decompress everything)
- More complex format (need to track boundaries)

### 3. Higher Compression Level (19)
```
Level 19, each binary compressed independently:
  Total: 5.00 MB compressed (58.2% ratio)

Improvement: 0.20 MB saved (3.7% better)
```

**Pros:**
- Slightly better compression
- Still allows individual extraction

**Cons:**
- **Minimal improvement** (only 3.7%)
- Much slower compression time
- Not worth the trade-off for fat binaries

## Recommendations

### For Current Use Case (Runtime Selection)

**Keep the current approach** (independent compression at level 3):

1. **Fast runtime selection**: Only decompress the binary you need
2. **Low memory usage**: Decompress ~1.4 MB instead of 3.5 MB
3. **Simple implementation**: No format changes needed
4. **Good enough**: 60% compression is already quite good

The 32% space savings from combined compression is **not worth** the runtime penalty of decompressing all binaries every time.

### Potential Future Optimizations

If space is critical (e.g., for distribution), consider:

#### 1. **Hybrid Approach: Separate Archives**
```
Create two formats:
- Standard .fat: Independent compression (runtime-friendly)
- Compressed .fat.zst: Combined compression (distribution)

Usage:
  # For distribution
  adipo create --compress-archive app.fat.zst ...

  # For deployment, expand once
  adipo expand app.fat.zst app.fat

  # Runtime uses standard format
  ./app.fat
```

**Benefits:**
- Best of both worlds
- Save bandwidth/storage for distribution
- Fast runtime after one-time expansion

#### 2. **Dictionary-Based Compression**
Use first binary as training data for shared dictionary:
- Dictionary size: ~110 KB
- Potential savings: 5-15% on subsequent binaries
- Complexity: Medium
- **Status**: Tested but less effective than combined stream

#### 3. **Delta Compression**
Store v1 fully, then deltas for v2-v4:
- Could achieve similar savings to combined stream
- More complex implementation
- Each binary would need base + delta to decompress
- **Status**: Not tested, potentially complex

## Conclusion

For `adipo`'s primary use case (runtime selection with self-extracting binaries):

✅ **Current approach is optimal**: Independent compression balances compression ratio with runtime performance

❌ **Don't use combined compression**: 32% space savings doesn't justify the runtime cost

✅ **Consider hybrid format**: If distribution size is critical, add optional archive format

❌ **Don't increase compression level**: Level 19 only saves 3.7% with major time cost

## Future Considerations

If binary sizes grow significantly (10+ binaries, or very large binaries):
1. Revisit combined compression for distribution-only format
2. Consider chunk-level deduplication
3. Explore incremental/streaming decompression techniques

## Test Data

```bash
# Reproduce these tests:
cd /tmp
cat > hello.go << 'EOF'
package main
import "fmt"
func main() { fmt.Println("Hello!") }
EOF

# Build binaries
GOOS=linux GOARCH=amd64 GOAMD64=v1 go build -o hello-v1 hello.go
GOOS=linux GOARCH=amd64 GOAMD64=v2 go build -o hello-v2 hello.go
GOOS=linux GOARCH=amd64 GOAMD64=v3 go build -o hello-v3 hello.go
GOOS=linux GOARCH=amd64 GOAMD64=v4 go build -o hello-v4 hello.go

# Test compression strategies (use adipo codebase for dependencies)
# See test_compression2.go in /tmp
```
