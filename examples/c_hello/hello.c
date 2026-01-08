#include <stdio.h>
#include <stdint.h>

#ifdef __AVX2__
#include <immintrin.h>
#endif

// Simple function that benefits from AVX2 if available
void compute_sum(int32_t *data, size_t len, int64_t *result) {
    int64_t sum = 0;

#ifdef __AVX2__
    // AVX2 version - processes 8 ints at a time
    __m256i vsum = _mm256_setzero_si256();
    size_t i = 0;

    for (; i + 8 <= len; i += 8) {
        __m256i vdata = _mm256_loadu_si256((__m256i*)(data + i));
        vsum = _mm256_add_epi32(vsum, vdata);
    }

    // Horizontal sum of the vector
    __m128i low = _mm256_castsi256_si128(vsum);
    __m128i high = _mm256_extracti128_si256(vsum, 1);
    __m128i sum128 = _mm_add_epi32(low, high);
    sum128 = _mm_hadd_epi32(sum128, sum128);
    sum128 = _mm_hadd_epi32(sum128, sum128);
    sum = _mm_extract_epi32(sum128, 0);

    // Process remaining elements
    for (; i < len; i++) {
        sum += data[i];
    }
#else
    // Scalar version
    for (size_t i = 0; i < len; i++) {
        sum += data[i];
    }
#endif

    *result = sum;
}

int main() {
    printf("Hello from C!\n");

#ifdef __AVX2__
    printf("Compiled with AVX2 support (x86-64-v3)\n");
#elif defined(__SSE4_2__)
    printf("Compiled with SSE4.2 support (x86-64-v2)\n");
#else
    printf("Compiled with baseline x86-64 (x86-64-v1)\n");
#endif

    // Test the compute_sum function
    int32_t data[1000];
    for (int i = 0; i < 1000; i++) {
        data[i] = i;
    }

    int64_t result;
    compute_sum(data, 1000, &result);
    printf("Sum of 0..999 = %lld (expected: 499500)\n", (long long)result);

    return 0;
}
