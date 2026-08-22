# Benchmark Summary

- Started: 2026-05-17T12:12:40Z
- Finished: 2026-05-17T18:28:00Z
- Layout(s): packed, split
- CPU: AMD Ryzen 9 9950X 16-Core Processor (32 logical cores)
- Memory: 60 GiB
- Kernel: Linux 7.0.5-arch1-1
- Go: go version go1.25.5 linux/amd64

## Single Thread Packed Basic Native Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | native | BRU_log2t2 | 1 | 10 | 4.9 ms | 228 us | 1 us | 29.2 bits | 27.7 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | BRU_log2t3 | 1 | 10 | 4.9 ms | 232 us | 2 us | 29.1 bits | 27.5 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | BRU_log2t4 | 1 | 10 | 4.9 ms | 256 us | 4 us | 29.1 bits | 27.7 bits | -1.4 bits | yes | 0 |
| basic | 15 | native | BRU_log2t5 | 1 | 10 | 4.9 ms | 260 us | 9 us | 29.1 bits | 27.4 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | BRU_log2t6 | 1 | 10 | 4.8 ms | 211 us | 19 us | 29.1 bits | 27.7 bits | -1.4 bits | yes | 0 |
| basic | 15 | native | BRU_log2t7 | 1 | 10 | 4.8 ms | 150 us | 37 us | 29.1 bits | 27.5 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | BRU_log2t8 | 1 | 10 | 4.8 ms | 198 us | 75 us | 29.2 bits | 27.5 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | BRU_log2t9 | 1 | 10 | 4.8 ms | 210 us | 152 us | 29.2 bits | 27.7 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | BRU_log2t10 | 1 | 10 | 4.9 ms | 246 us | 305 us | 29.0 bits | 27.4 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t2 | 1 | 10 | 4.9 ms | 229 us | 1 us | 29.2 bits | 27.8 bits | -1.3 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t3 | 1 | 10 | 4.9 ms | 272 us | 2 us | 29.1 bits | 27.9 bits | -1.2 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t4 | 1 | 10 | 4.9 ms | 265 us | 4 us | 29.2 bits | 27.6 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t5 | 1 | 10 | 4.8 ms | 99 us | 9 us | 29.2 bits | 27.8 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t6 | 1 | 10 | 4.8 ms | 167 us | 17 us | 29.2 bits | 27.8 bits | -1.4 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t7 | 1 | 10 | 7.0 ms | 1.1 ms | 54 us | 29.1 bits | 27.7 bits | -1.3 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t8 | 1 | 10 | 5.6 ms | 1.2 ms | 87 us | 29.1 bits | 26.8 bits | -2.3 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t9 | 1 | 10 | 4.8 ms | 181 us | 150 us | 29.2 bits | 27.7 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t10 | 1 | 10 | 4.8 ms | 198 us | 300 us | 29.2 bits | 27.6 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | WH_log2t2 | 1 | 10 | 4.8 ms | 181 us | 1 us | 29.2 bits | 27.4 bits | -1.8 bits | yes | 0 |
| basic | 15 | native | WH_log2t3 | 1 | 10 | 4.8 ms | 183 us | 2 us | 29.2 bits | 27.6 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | WH_log2t4 | 1 | 10 | 4.7 ms | 15 us | 4 us | 29.2 bits | 27.7 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | WH_log2t5 | 1 | 10 | 4.7 ms | 25 us | 9 us | 29.2 bits | 27.8 bits | -1.3 bits | yes | 0 |
| basic | 15 | native | WH_log2t6 | 1 | 10 | 4.7 ms | 169 us | 18 us | 29.1 bits | 27.5 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | WH_log2t7 | 1 | 10 | 4.8 ms | 209 us | 37 us | 29.2 bits | 27.7 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | WH_log2t8 | 1 | 10 | 4.7 ms | 146 us | 74 us | 29.1 bits | 27.8 bits | -1.4 bits | yes | 0 |
| basic | 15 | native | WH_log2t9 | 1 | 10 | 4.7 ms | 144 us | 148 us | 29.1 bits | 27.8 bits | -1.2 bits | yes | 0 |
| basic | 15 | native | WH_log2t10 | 1 | 10 | 4.8 ms | 175 us | 297 us | 29.1 bits | 27.6 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | IND_log2t2 | 1 | 10 | 4.8 ms | 197 us | 1 us | 29.2 bits | 27.7 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | IND_log2t3 | 1 | 10 | 4.7 ms | 34 us | 2 us | 29.3 bits | 27.7 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | IND_log2t4 | 1 | 10 | 4.7 ms | 21 us | 4 us | 29.2 bits | 27.3 bits | -1.9 bits | yes | 0 |
| basic | 15 | native | IND_log2t5 | 1 | 10 | 4.8 ms | 218 us | 9 us | 29.2 bits | 27.9 bits | -1.3 bits | yes | 0 |
| basic | 15 | native | IND_log2t6 | 1 | 10 | 4.8 ms | 211 us | 18 us | 29.1 bits | 27.7 bits | -1.4 bits | yes | 0 |
| basic | 15 | native | IND_log2t7 | 1 | 10 | 4.7 ms | 57 us | 37 us | 29.1 bits | 27.3 bits | -1.9 bits | yes | 0 |
| basic | 15 | native | IND_log2t8 | 1 | 10 | 4.8 ms | 165 us | 75 us | 29.2 bits | 27.8 bits | -1.3 bits | yes | 0 |
| basic | 15 | native | IND_log2t9 | 1 | 10 | 4.8 ms | 162 us | 149 us | 29.1 bits | 27.8 bits | -1.3 bits | yes | 0 |
| basic | 15 | native | IND_log2t10 | 1 | 10 | 4.8 ms | 166 us | 297 us | 29.0 bits | 27.9 bits | -1.2 bits | yes | 0 |

## Single Thread Split Basic Native Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | native | BRU_log2t2 | 1 | 10 | 14.0 ms | 38 us | 1 us | 29.0 bits | 27.7 bits | -1.2 bits | yes | 0 |
| basic | 15 | native | BRU_log2t3 | 1 | 10 | 32.9 ms | 103 us | 2 us | 28.8 bits | 27.4 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | BRU_log2t4 | 1 | 10 | 71.1 ms | 397 us | 4 us | 29.0 bits | 27.3 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | BRU_log2t5 | 1 | 10 | 145.4 ms | 360 us | 9 us | 29.0 bits | 27.3 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | BRU_log2t6 | 1 | 10 | 296.0 ms | 1.3 ms | 18 us | 28.9 bits | 27.3 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | BRU_log2t7 | 1 | 10 | 595.1 ms | 1.1 ms | 36 us | 28.8 bits | 27.2 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | BRU_log2t8 | 1 | 10 | 1.20 s | 8.1 ms | 73 us | 28.8 bits | 27.2 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t2 | 1 | 10 | 9.6 ms | 209 us | 1 us | 29.1 bits | 27.4 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t3 | 1 | 10 | 28.7 ms | 218 us | 2 us | 28.9 bits | 27.3 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t4 | 1 | 10 | 57.9 ms | 505 us | 4 us | 29.1 bits | 27.3 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t5 | 1 | 10 | 141.7 ms | 487 us | 9 us | 28.9 bits | 27.3 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t6 | 1 | 10 | 284.6 ms | 1.0 ms | 17 us | 28.9 bits | 27.0 bits | -1.9 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t7 | 1 | 10 | 591.5 ms | 2.1 ms | 36 us | 28.7 bits | 27.3 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t8 | 1 | 10 | 1.18 s | 7.6 ms | 72 us | 28.8 bits | 27.2 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | WH_log2t2 | 1 | 10 | 14.3 ms | 81 us | 1 us | 29.1 bits | 27.4 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | WH_log2t3 | 1 | 10 | 33.2 ms | 151 us | 2 us | 28.9 bits | 27.4 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | WH_log2t4 | 1 | 10 | 71.9 ms | 362 us | 4 us | 29.0 bits | 27.4 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | WH_log2t5 | 1 | 10 | 147.0 ms | 423 us | 9 us | 29.0 bits | 27.5 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | WH_log2t6 | 1 | 10 | 296.6 ms | 1.5 ms | 18 us | 28.9 bits | 27.1 bits | -1.9 bits | yes | 0 |
| basic | 15 | native | WH_log2t7 | 1 | 10 | 594.2 ms | 1.4 ms | 36 us | 28.8 bits | 27.2 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | WH_log2t8 | 1 | 10 | 1.19 s | 3.3 ms | 73 us | 28.8 bits | 27.2 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | IND_log2t2 | 1 | 10 | 14.1 ms | 104 us | 1 us | 28.9 bits | 27.7 bits | -1.2 bits | yes | 0 |
| basic | 15 | native | IND_log2t3 | 1 | 10 | 32.9 ms | 84 us | 2 us | 29.0 bits | 27.5 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | IND_log2t4 | 1 | 10 | 71.3 ms | 639 us | 4 us | 29.0 bits | 27.2 bits | -1.8 bits | yes | 0 |
| basic | 15 | native | IND_log2t5 | 1 | 10 | 145.7 ms | 412 us | 9 us | 29.0 bits | 27.3 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | IND_log2t6 | 1 | 10 | 294.4 ms | 713 us | 18 us | 29.0 bits | 27.3 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | IND_log2t7 | 1 | 10 | 594.3 ms | 2.0 ms | 36 us | 28.8 bits | 26.9 bits | -1.9 bits | yes | 0 |
| basic | 15 | native | IND_log2t8 | 1 | 10 | 1.19 s | 2.1 ms | 73 us | 28.9 bits | 27.2 bits | -1.7 bits | yes | 0 |

## Single Thread Packed Basic Unary LUT Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | unary_lut | BRU_log2t2 | 1 | 10 | 8.7 ms | 21 us | 2 us | 29.3 bits | 27.4 bits | -1.9 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t3 | 1 | 10 | 15.8 ms | 40 us | 7 us | 29.1 bits | 27.5 bits | -1.6 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t4 | 1 | 10 | 24.2 ms | 261 us | 22 us | 29.1 bits | 27.6 bits | -1.6 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t5 | 1 | 10 | 34.9 ms | 1.4 ms | 66 us | 29.1 bits | 27.6 bits | -1.5 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t6 | 1 | 10 | 53.5 ms | 280 us | 206 us | 29.1 bits | 27.5 bits | -1.5 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t7 | 1 | 10 | 80.1 ms | 2.7 ms | 621 us | 29.2 bits | 27.4 bits | -1.9 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t8 | 1 | 10 | 122.5 ms | 608 us | 1.9 ms | 29.2 bits | 27.3 bits | -1.9 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t9 | 1 | 10 | 203.8 ms | 3.2 ms | 6.4 ms | 29.1 bits | 27.0 bits | -2.1 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t10 | 1 | 10 | 342.3 ms | 1.1 ms | 21.4 ms | 29.0 bits | 26.9 bits | -2.1 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t2 | 1 | 10 | 7.7 ms | 232 us | 1 us | 29.1 bits | 27.6 bits | -1.5 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t3 | 1 | 10 | 15.5 ms | 51 us | 6 us | 29.1 bits | 27.3 bits | -1.8 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t4 | 1 | 10 | 20.5 ms | 348 us | 15 us | 29.2 bits | 26.8 bits | -2.4 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t5 | 1 | 10 | 33.4 ms | 60 us | 61 us | 29.2 bits | 27.1 bits | -2.0 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t6 | 1 | 10 | 50.0 ms | 325 us | 183 us | 28.9 bits | 26.9 bits | -2.1 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t7 | 1 | 10 | 78.0 ms | 142 us | 600 us | 29.1 bits | 27.2 bits | -1.9 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t8 | 1 | 10 | 124.3 ms | 534 us | 1.9 ms | 29.2 bits | 27.1 bits | -2.1 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t9 | 1 | 10 | 202.0 ms | 890 us | 6.3 ms | 29.1 bits | 27.1 bits | -2.0 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t10 | 1 | 10 | 354.5 ms | 4.5 ms | 22.2 ms | 28.9 bits | 26.9 bits | -2.1 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t2 | 1 | 10 | 8.8 ms | 28 us | 2 us | 29.2 bits | 27.4 bits | -1.8 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t3 | 1 | 10 | 15.8 ms | 66 us | 7 us | 29.2 bits | 27.3 bits | -1.9 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t4 | 1 | 10 | 24.2 ms | 50 us | 22 us | 29.2 bits | 27.5 bits | -1.7 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t5 | 1 | 10 | 34.0 ms | 76 us | 64 us | 29.1 bits | 27.6 bits | -1.4 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t6 | 1 | 10 | 54.0 ms | 239 us | 208 us | 29.2 bits | 27.3 bits | -1.9 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t7 | 1 | 10 | 78.6 ms | 685 us | 609 us | 29.1 bits | 27.4 bits | -1.7 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t8 | 1 | 10 | 124.8 ms | 5.8 ms | 1.9 ms | 29.2 bits | 27.3 bits | -2.0 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t9 | 1 | 10 | 203.2 ms | 2.9 ms | 6.4 ms | 29.3 bits | 27.1 bits | -2.2 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t10 | 1 | 10 | 350.3 ms | 20.0 ms | 21.9 ms | 29.2 bits | 26.4 bits | -2.8 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t2 | 1 | 10 | 8.6 ms | 51 us | 2 us | 29.2 bits | 27.5 bits | -1.7 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t3 | 1 | 10 | 12.3 ms | 28 us | 5 us | 29.2 bits | 26.9 bits | -2.3 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t4 | 1 | 10 | 19.8 ms | 136 us | 18 us | 29.2 bits | 26.2 bits | -3.0 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t5 | 1 | 10 | 28.4 ms | 138 us | 54 us | 29.2 bits | 26.2 bits | -3.0 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t6 | 1 | 10 | 45.9 ms | 168 us | 176 us | 29.1 bits | 25.5 bits | -3.7 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t7 | 1 | 10 | 69.9 ms | 327 us | 542 us | 29.1 bits | 25.1 bits | -4.0 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t8 | 1 | 10 | 107.4 ms | 560 us | 1.7 ms | 29.1 bits | 24.7 bits | -4.4 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t9 | 1 | 10 | 172.8 ms | 1.0 ms | 5.4 ms | 29.1 bits | 24.5 bits | -4.6 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t10 | 1 | 10 | 271.6 ms | 2.0 ms | 17.0 ms | 29.1 bits | 24.2 bits | -4.9 bits | yes | 0 |

## Single Thread Split Basic Unary LUT Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | unary_lut | BRU_log2t2 | 1 | 10 | 4.0 ms | 258 us | 0 us | 29.1 bits | 27.8 bits | -1.3 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t3 | 1 | 10 | 14.2 ms | 350 us | 1 us | 29.1 bits | 27.5 bits | -1.6 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t4 | 1 | 10 | 47.1 ms | 382 us | 3 us | 28.9 bits | 27.3 bits | -1.6 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t5 | 1 | 10 | 168.8 ms | 806 us | 10 us | 29.0 bits | 27.3 bits | -1.6 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t6 | 1 | 10 | 647.9 ms | 5.5 ms | 40 us | 28.9 bits | 27.3 bits | -1.7 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t7 | 1 | 10 | 2.48 s | 6.4 ms | 151 us | 28.8 bits | 27.3 bits | -1.5 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t8 | 1 | 10 | 9.72 s | 69.6 ms | 593 us | 28.9 bits | 27.3 bits | -1.6 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t2 | 1 | 10 | 2.0 ms | 240 us | 0 us | 29.1 bits | 29.1 bits | 0.0 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t3 | 1 | 10 | 10.4 ms | 355 us | 1 us | 29.1 bits | 27.6 bits | -1.5 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t4 | 1 | 10 | 31.3 ms | 619 us | 2 us | 29.0 bits | 27.2 bits | -1.8 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t5 | 1 | 10 | 157.4 ms | 1.3 ms | 10 us | 29.0 bits | 27.3 bits | -1.7 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t6 | 1 | 10 | 587.3 ms | 5.3 ms | 36 us | 28.8 bits | 27.1 bits | -1.8 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t7 | 1 | 10 | 2.46 s | 5.6 ms | 150 us | 28.9 bits | 27.3 bits | -1.6 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t8 | 1 | 10 | 9.35 s | 35.7 ms | 571 us | 28.8 bits | 27.1 bits | -1.7 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t2 | 1 | 10 | 3.5 ms | 35 us | 0 us | 29.0 bits | 27.8 bits | -1.2 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t3 | 1 | 10 | 11.9 ms | 103 us | 1 us | 29.0 bits | 27.5 bits | -1.5 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t4 | 1 | 10 | 40.5 ms | 402 us | 2 us | 28.9 bits | 27.4 bits | -1.4 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t5 | 1 | 10 | 151.0 ms | 983 us | 9 us | 29.0 bits | 27.5 bits | -1.5 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t6 | 1 | 10 | 583.0 ms | 5.4 ms | 36 us | 28.9 bits | 27.2 bits | -1.7 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t7 | 1 | 10 | 2.32 s | 6.4 ms | 141 us | 28.9 bits | 27.3 bits | -1.6 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t8 | 1 | 10 | 9.19 s | 18.5 ms | 561 us | 28.8 bits | 27.1 bits | -1.7 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t2 | 1 | 10 | 3.3 ms | 54 us | 0 us | 29.1 bits | 28.4 bits | -0.7 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t3 | 1 | 10 | 7.7 ms | 89 us | 0 us | 28.9 bits | 27.7 bits | -1.2 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t4 | 1 | 10 | 17.7 ms | 166 us | 1 us | 28.9 bits | 27.1 bits | -1.8 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t5 | 1 | 10 | 36.3 ms | 243 us | 2 us | 28.9 bits | 26.8 bits | -2.2 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t6 | 1 | 10 | 73.2 ms | 552 us | 4 us | 28.8 bits | 26.2 bits | -2.6 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t7 | 1 | 10 | 148.5 ms | 640 us | 9 us | 28.8 bits | 25.7 bits | -3.1 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t8 | 1 | 10 | 293.4 ms | 2.2 ms | 18 us | 28.8 bits | 25.1 bits | -3.8 bits | yes | 0 |

## Single Thread Packed Basic Bivariate LUT Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | binary_lut | BRU_log2t2 | 3 | 10 | 102.1 ms | 338 us | 100 us | 29.4 bits | 27.3 bits | -2.1 bits | yes | 0 |
| basic | 15 | binary_lut | BRU_log2t3 | 3 | 10 | 170.5 ms | 777 us | 666 us | 29.4 bits | 27.2 bits | -2.2 bits | yes | 0 |
| basic | 15 | binary_lut | BRU_log2t4 | 3 | 10 | 324.2 ms | 1.0 ms | 5.1 ms | 29.2 bits | 27.5 bits | -1.8 bits | yes | 0 |
| basic | 15 | binary_lut | BRU_log2t5 | 3 | 10 | 725.0 ms | 3.0 ms | 45.3 ms | 29.5 bits | 27.6 bits | -1.9 bits | yes | 0 |
| basic | 15 | binary_lut | LBRU_log2t2 | 3 | 10 | 76.9 ms | 204 us | 42 us | 29.2 bits | 26.3 bits | -2.9 bits | yes | 0 |
| basic | 15 | binary_lut | LBRU_log2t3 | 3 | 10 | 149.6 ms | 3.3 ms | 448 us | 29.3 bits | 25.9 bits | -3.4 bits | yes | 0 |
| basic | 15 | binary_lut | LBRU_log2t4 | 3 | 10 | 259.2 ms | 426 us | 2.7 ms | 29.3 bits | 26.2 bits | -3.1 bits | yes | 0 |
| basic | 15 | binary_lut | LBRU_log2t5 | 3 | 10 | 693.9 ms | 4.0 ms | 40.8 ms | 29.5 bits | 26.2 bits | -3.3 bits | yes | 0 |
| basic | 15 | binary_lut | WH_log2t2 | 3 | 10 | 102.4 ms | 761 us | 100 us | 29.3 bits | 26.9 bits | -2.5 bits | yes | 0 |
| basic | 15 | binary_lut | WH_log2t3 | 3 | 10 | 171.5 ms | 2.1 ms | 670 us | 29.2 bits | 27.4 bits | -1.8 bits | yes | 0 |
| basic | 15 | binary_lut | WH_log2t4 | 3 | 10 | 323.6 ms | 1.3 ms | 5.1 ms | 29.2 bits | 27.3 bits | -1.9 bits | yes | 0 |
| basic | 15 | binary_lut | WH_log2t5 | 3 | 10 | 722.1 ms | 4.1 ms | 45.1 ms | 29.5 bits | 27.1 bits | -2.4 bits | yes | 0 |
| basic | 15 | binary_lut | IND_log2t2 | 3 | 10 | 98.3 ms | 137 us | 96 us | 29.3 bits | 25.8 bits | -3.5 bits | yes | 0 |
| basic | 15 | binary_lut | IND_log2t3 | 3 | 10 | 168.2 ms | 684 us | 657 us | 29.4 bits | 24.9 bits | -4.5 bits | yes | 0 |
| basic | 15 | binary_lut | IND_log2t4 | 3 | 10 | 322.7 ms | 2.4 ms | 5.0 ms | 29.4 bits | 24.5 bits | -4.8 bits | yes | 0 |
| basic | 15 | binary_lut | IND_log2t5 | 3 | 10 | 721.8 ms | 5.7 ms | 45.1 ms | 29.3 bits | 24.1 bits | -5.2 bits | yes | 0 |

## Single Thread Split Basic Bivariate LUT Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | binary_lut | BRU_log2t2 | 2 | 10 | 79.0 ms | 566 us | 5 us | 29.1 bits | 19.1 bits | -10.0 bits | yes | 0 |
| basic | 15 | binary_lut | BRU_log2t3 | 2 | 10 | 445.5 ms | 1.5 ms | 27 us | 29.0 bits | 19.3 bits | -9.7 bits | yes | 0 |
| basic | 15 | binary_lut | BRU_log2t4 | 2 | 10 | 2.28 s | 6.8 ms | 139 us | 29.0 bits | 19.5 bits | -9.5 bits | yes | 0 |
| basic | 15 | binary_lut | BRU_log2t5 | 2 | 10 | 11.84 s | 45.8 ms | 722 us | 29.0 bits | 19.7 bits | -9.3 bits | yes | 0 |
| basic | 15 | binary_lut | LBRU_log2t2 | 2 | 10 | 33.8 ms | 241 us | 2 us | 29.1 bits | 18.4 bits | -10.7 bits | yes | 0 |
| basic | 15 | binary_lut | LBRU_log2t3 | 2 | 10 | 323.7 ms | 525 us | 20 us | 29.0 bits | 17.6 bits | -11.5 bits | yes | 0 |
| basic | 15 | binary_lut | LBRU_log2t4 | 2 | 10 | 1.41 s | 6.3 ms | 86 us | 28.9 bits | 17.4 bits | -11.5 bits | yes | 0 |
| basic | 15 | binary_lut | LBRU_log2t5 | 2 | 10 | 11.02 s | 36.8 ms | 673 us | 28.9 bits | 17.4 bits | -11.5 bits | yes | 0 |
| basic | 15 | binary_lut | WH_log2t2 | 2 | 10 | 78.0 ms | 727 us | 5 us | 29.1 bits | 18.6 bits | -10.5 bits | yes | 0 |
| basic | 15 | binary_lut | WH_log2t3 | 2 | 10 | 440.6 ms | 1.7 ms | 27 us | 29.0 bits | 19.1 bits | -9.9 bits | yes | 0 |
| basic | 15 | binary_lut | WH_log2t4 | 2 | 10 | 2.26 s | 4.4 ms | 138 us | 29.0 bits | 19.0 bits | -10.0 bits | yes | 0 |
| basic | 15 | binary_lut | WH_log2t5 | 2 | 10 | 11.81 s | 46.2 ms | 721 us | 29.0 bits | 19.3 bits | -9.7 bits | yes | 0 |
| basic | 15 | binary_lut | IND_log2t2 | 2 | 10 | 75.6 ms | 313 us | 5 us | 29.1 bits | 18.4 bits | -10.7 bits | yes | 0 |
| basic | 15 | binary_lut | IND_log2t3 | 2 | 10 | 403.3 ms | 798 us | 25 us | 29.0 bits | 18.4 bits | -10.6 bits | yes | 0 |
| basic | 15 | binary_lut | IND_log2t4 | 2 | 10 | 1.84 s | 4.2 ms | 112 us | 29.0 bits | 18.4 bits | -10.6 bits | yes | 0 |
| basic | 15 | binary_lut | IND_log2t5 | 2 | 10 | 7.81 s | 30.8 ms | 477 us | 29.0 bits | 18.4 bits | -10.6 bits | yes | 0 |

## Single Thread Basic 4-Variate LUT Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | four_lut | BRU_log2t2 | 4 | 10 | 943.5 ms | 3.8 ms | 14.7 ms | 29.2 bits | 26.8 bits | -2.4 bits | yes | 0 |
| basic | 15 | four_lut | BRU_log2t3 | 4 | 10 | 5.21 s | 24.5 ms | 1.30 s | 29.9 bits | 26.9 bits | -3.0 bits | yes | 0 |
| basic | 15 | four_lut | LBRU_log2t2 | 4 | 10 | 511.0 ms | 2.3 ms | 2.5 ms | 29.4 bits | 24.6 bits | -4.8 bits | yes | 0 |
| basic | 15 | four_lut | LBRU_log2t3 | 4 | 10 | 3.61 s | 14.0 ms | 602.0 ms | 29.9 bits | 24.8 bits | -5.1 bits | yes | 0 |
| basic | 15 | four_lut | WH_log2t2 | 4 | 10 | 943.6 ms | 2.6 ms | 14.7 ms | 29.7 bits | 27.0 bits | -2.7 bits | yes | 0 |
| basic | 15 | four_lut | WH_log2t3 | 4 | 10 | 5.22 s | 31.3 ms | 1.31 s | 29.9 bits | 27.3 bits | -2.7 bits | yes | 0 |
| basic | 15 | four_lut | IND_log2t2 | 4 | 10 | 939.8 ms | 1.5 ms | 14.7 ms | 29.5 bits | 23.8 bits | -5.7 bits | yes | 0 |
| basic | 15 | four_lut | IND_log2t3 | 4 | 10 | 5.12 s | 8.4 ms | 1.28 s | 29.9 bits | 22.5 bits | -7.4 bits | yes | 0 |

## Single Thread Packed Basic Clean Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | clean | BRU_log2t2 | 4 | 10 | 56.4 ms | 106 us | 10 us | 11.8 bits | 16.5 bits | +4.7 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t3 | 4 | 10 | 87.9 ms | 293 us | 38 us | 11.8 bits | 16.5 bits | +4.7 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t4 | 4 | 10 | 123.6 ms | 294 us | 113 us | 11.8 bits | 16.5 bits | +4.7 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t5 | 4 | 10 | 163.1 ms | 462 us | 309 us | 11.8 bits | 16.5 bits | +4.7 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t6 | 4 | 10 | 243.4 ms | 549 us | 936 us | 11.8 bits | 16.5 bits | +4.7 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t7 | 4 | 10 | 338.9 ms | 256 us | 2.6 ms | 11.8 bits | 16.5 bits | +4.7 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t8 | 4 | 10 | 507.9 ms | 6.1 ms | 7.9 ms | 11.8 bits | 16.5 bits | +4.7 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t9 | 4 | 10 | 801.2 ms | 7.1 ms | 25.0 ms | 11.8 bits | 16.5 bits | +4.7 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t10 | 4 | 10 | 1.28 s | 5.9 ms | 80.1 ms | 11.8 bits | 16.5 bits | +4.6 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t2 | 4 | 10 | 54.6 ms | 151 us | 7 us | 11.8 bits | 17.8 bits | +6.0 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t3 | 4 | 10 | 88.0 ms | 348 us | 32 us | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t4 | 4 | 10 | 109.0 ms | 284 us | 80 us | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t5 | 4 | 10 | 164.2 ms | 443 us | 301 us | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t6 | 4 | 10 | 229.8 ms | 471 us | 842 us | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t7 | 4 | 10 | 342.3 ms | 1.1 ms | 2.6 ms | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t8 | 4 | 10 | 522.0 ms | 5.4 ms | 8.0 ms | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t9 | 4 | 10 | 801.5 ms | 8.6 ms | 25.0 ms | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t10 | 4 | 10 | 1.29 s | 10.4 ms | 80.9 ms | 11.8 bits | 17.7 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | WH_log2t2 | 2 | 10 | 12.6 ms | 115 us | 2 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t3 | 2 | 10 | 12.5 ms | 106 us | 5 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t4 | 2 | 10 | 12.5 ms | 95 us | 11 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t5 | 2 | 10 | 12.5 ms | 50 us | 24 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t6 | 2 | 10 | 12.5 ms | 75 us | 48 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t7 | 2 | 10 | 12.4 ms | 37 us | 96 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t8 | 2 | 10 | 12.5 ms | 89 us | 195 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t9 | 2 | 10 | 12.5 ms | 78 us | 389 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t10 | 2 | 10 | 12.5 ms | 73 us | 780 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | IND_log2t2 | 2 | 10 | 12.5 ms | 61 us | 2 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t3 | 2 | 10 | 12.6 ms | 249 us | 5 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t4 | 2 | 10 | 12.6 ms | 128 us | 12 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t5 | 2 | 10 | 12.5 ms | 57 us | 24 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t6 | 2 | 10 | 12.5 ms | 58 us | 48 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t7 | 2 | 10 | 12.6 ms | 234 us | 98 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t8 | 2 | 10 | 12.6 ms | 234 us | 197 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t9 | 2 | 10 | 12.5 ms | 70 us | 391 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t10 | 2 | 10 | 12.5 ms | 105 us | 783 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |

## Single Thread Split Basic Clean Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | clean | BRU_log2t2 | 4 | 10 | 71.2 ms | 389 us | 4 us | 11.8 bits | 16.8 bits | +5.0 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t3 | 4 | 10 | 179.3 ms | 1.3 ms | 11 us | 11.8 bits | 16.8 bits | +5.0 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t4 | 4 | 10 | 450.0 ms | 3.3 ms | 27 us | 11.8 bits | 16.8 bits | +5.0 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t5 | 4 | 10 | 1.19 s | 6.8 ms | 72 us | 11.8 bits | 16.8 bits | +5.0 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t6 | 4 | 10 | 3.41 s | 15.2 ms | 208 us | 11.8 bits | 16.8 bits | +5.0 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t7 | 4 | 10 | 10.93 s | 36.8 ms | 667 us | 11.8 bits | 16.8 bits | +4.9 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t8 | 4 | 10 | 38.39 s | 408.2 ms | 2.3 ms | 11.8 bits | 16.8 bits | +4.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t2 | 4 | 10 | 47.0 ms | 385 us | 3 us | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t3 | 4 | 10 | 154.3 ms | 1.9 ms | 9 us | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t4 | 4 | 10 | 342.4 ms | 7.4 ms | 21 us | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t5 | 4 | 10 | 1.12 s | 3.7 ms | 68 us | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t6 | 4 | 10 | 3.16 s | 6.6 ms | 193 us | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t7 | 4 | 10 | 10.80 s | 24.2 ms | 659 us | 11.8 bits | 17.7 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t8 | 4 | 10 | 36.84 s | 73.6 ms | 2.2 ms | 11.8 bits | 17.7 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | WH_log2t2 | 2 | 10 | 37.7 ms | 203 us | 2 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t3 | 2 | 10 | 89.5 ms | 1.2 ms | 5 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t4 | 2 | 10 | 191.8 ms | 4.6 ms | 12 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t5 | 2 | 10 | 388.7 ms | 860 us | 24 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t6 | 2 | 10 | 790.1 ms | 3.5 ms | 48 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t7 | 2 | 10 | 1.59 s | 4.3 ms | 97 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t8 | 2 | 10 | 3.24 s | 91.4 ms | 198 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | IND_log2t2 | 2 | 10 | 38.0 ms | 156 us | 2 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t3 | 2 | 10 | 89.7 ms | 393 us | 5 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t4 | 2 | 10 | 190.6 ms | 775 us | 12 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t5 | 2 | 10 | 392.3 ms | 1.2 ms | 24 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t6 | 2 | 10 | 791.4 ms | 4.8 ms | 48 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t7 | 2 | 10 | 1.59 s | 3.0 ms | 97 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t8 | 2 | 10 | 3.20 s | 21.4 ms | 195 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |

## Single Thread Basic Split->Standard Results

| Target | LogN | Operation | Shape | Levels | Remaining | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | --------: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | to_standard | BRU_log2t2 | 1 | 0 | 10 | 1.4 ms | 59 us | 0 us |  |  |  | yes | 0 |
| basic | 15 | to_standard | BRU_log2t3 | 1 | 0 | 10 | 2.4 ms | 109 us | 0 us |  |  |  | yes | 0 |
| basic | 15 | to_standard | BRU_log2t4 | 1 | 0 | 10 | 4.1 ms | 241 us | 0 us |  |  |  | yes | 0 |
| basic | 15 | to_standard | BRU_log2t5 | 1 | 0 | 10 | 7.5 ms | 252 us | 0 us |  |  |  | yes | 0 |
| basic | 15 | to_standard | BRU_log2t6 | 1 | 0 | 10 | 13.9 ms | 141 us | 1 us |  |  |  | yes | 0 |
| basic | 15 | to_standard | BRU_log2t7 | 1 | 0 | 10 | 27.3 ms | 658 us | 2 us |  |  |  | yes | 0 |
| basic | 15 | to_standard | BRU_log2t8 | 1 | 0 | 10 | 54.9 ms | 1.1 ms | 3 us |  |  |  | yes | 0 |

## Single Thread Basic Standard->Split Results

| Target | LogN | Operation | Shape | Levels | Remaining | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | --------: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | from_standard | BRU_log2t2 | 9 | 6 | 10 | 5.16 s | 13.9 ms | 315 us |  |  |  | yes | 0 |
| basic | 15 | from_standard | BRU_log2t3 | 9 | 6 | 10 | 8.29 s | 49.7 ms | 506 us |  |  |  | yes | 0 |
| basic | 15 | from_standard | BRU_log2t4 | 9 | 6 | 10 | 14.62 s | 61.1 ms | 892 us |  |  |  | yes | 0 |
| basic | 15 | from_standard | BRU_log2t5 | 10 | 5 | 10 | 15.61 s | 95.6 ms | 953 us |  |  |  | yes | 0 |
| basic | 15 | from_standard | BRU_log2t6 | 12 | 3 | 10 | 16.13 s | 243.7 ms | 985 us |  |  |  | yes | 0 |
| basic | 15 | from_standard | BRU_log2t7 | 13 | 2 | 10 | 17.36 s | 951.0 ms | 1.1 ms |  |  |  | yes | 0 |

## Single Thread CRT Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | :------ | --------: |
| crt | 15 | add | 64bits | 1 | 10 | 4.7 ms | 23 us | 106 us | yes | 0 |
| crt | 15 | add | 256bits | 1 | 10 | 4.7 ms | 22 us | 1.2 ms | yes | 0 |
| crt | 15 | sub | 64bits | 1 | 10 | 8.4 ms | 61 us | 190 us | yes | 0 |
| crt | 15 | sub | 256bits | 1 | 10 | 8.4 ms | 71 us | 2.1 ms | yes | 0 |
| crt | 15 | mul_lbru | 64bits | 1 | 10 | 4.7 ms | 24 us | 106 us | yes | 0 |
| crt | 15 | mul_lbru | 256bits | 1 | 10 | 4.7 ms | 31 us | 1.2 ms | yes | 0 |
| crt | 15 | bru_to_lbru | 64bits | 1 | 10 | 48.1 ms | 244 us | 1.1 ms | yes | 0 |
| crt | 15 | bru_to_lbru | 256bits | 1 | 10 | 103.9 ms | 490 us | 26.0 ms | yes | 0 |
| crt | 15 | lbru_to_bru | 64bits | 1 | 10 | 48.2 ms | 155 us | 1.1 ms | yes | 0 |
| crt | 15 | lbru_to_bru | 256bits | 1 | 10 | 105.9 ms | 1.5 ms | 26.5 ms | yes | 0 |
| crt | 15 | clean | 64bits | 4 | 10 | 224.9 ms | 1.1 ms | 5.1 ms | yes | 0 |
| crt | 15 | clean | 256bits | 4 | 10 | 449.1 ms | 1.4 ms | 112.3 ms | yes | 0 |

## Single Thread Radix Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | :------ | --------: |
| radix | 16 | Add | 64bit_r4 | 9 | 10 | 2.11 s | 4.1 ms | 50.2 ms | yes | 0 |
| radix | 16 | Add | 256bit_r4 | 11 | 10 | 3.16 s | 17.9 ms | 316.1 ms | yes | 0 |
| radix | 16 | Sub | 64bit_r4 | 9 | 10 | 2.21 s | 11.8 ms | 52.7 ms | yes | 0 |
| radix | 16 | Sub | 256bit_r4 | 11 | 10 | 3.29 s | 16.8 ms | 328.5 ms | yes | 0 |
| radix | 16 | Eq | 64bit_r4 | 8 | 10 | 1.02 s | 3.2 ms | 24.4 ms | yes | 0 |
| radix | 16 | Eq | 256bit_r4 | 10 | 10 | 1.57 s | 6.7 ms | 156.5 ms | yes | 0 |
| radix | 16 | Lt | 64bit_r4 | 8 | 10 | 1.40 s | 2.3 ms | 33.2 ms | yes | 0 |
| radix | 16 | Lt | 256bit_r4 | 10 | 10 | 2.21 s | 6.6 ms | 220.6 ms | yes | 0 |
| radix | 16 | Cmp | 64bit_r4 | 8 | 10 | 1.78 s | 2.5 ms | 42.5 ms | yes | 0 |
| radix | 16 | Cmp | 256bit_r4 | 10 | 10 | 2.87 s | 12.4 ms | 286.8 ms | yes | 0 |
| radix | 16 | Clean | 64bit_r4 | 4 | 10 | 119.9 ms | 918 us | 2.9 ms | yes | 0 |
| radix | 16 | Clean | 256bit_r4 | 4 | 10 | 118.2 ms | 476 us | 11.8 ms | yes | 0 |

## Multi Thread Packed Basic Native Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | native | BRU_log2t2 | 1 | 10 | 4.9 ms | 173 us | 1 us | 29.2 bits | 27.8 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | BRU_log2t3 | 1 | 10 | 4.8 ms | 228 us | 2 us | 29.2 bits | 27.7 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | BRU_log2t4 | 1 | 10 | 4.9 ms | 148 us | 5 us | 29.1 bits | 27.7 bits | -1.4 bits | yes | 0 |
| basic | 15 | native | BRU_log2t5 | 1 | 10 | 4.8 ms | 82 us | 9 us | 29.0 bits | 27.8 bits | -1.3 bits | yes | 0 |
| basic | 15 | native | BRU_log2t6 | 1 | 10 | 4.9 ms | 158 us | 19 us | 29.1 bits | 27.7 bits | -1.4 bits | yes | 0 |
| basic | 15 | native | BRU_log2t7 | 1 | 10 | 4.8 ms | 153 us | 37 us | 29.2 bits | 27.7 bits | -1.4 bits | yes | 0 |
| basic | 15 | native | BRU_log2t8 | 1 | 10 | 4.8 ms | 138 us | 75 us | 29.1 bits | 27.6 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | BRU_log2t9 | 1 | 10 | 4.8 ms | 124 us | 150 us | 29.2 bits | 27.6 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | BRU_log2t10 | 1 | 10 | 4.9 ms | 230 us | 306 us | 29.2 bits | 27.8 bits | -1.4 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t2 | 1 | 10 | 4.8 ms | 106 us | 1 us | 29.1 bits | 27.7 bits | -1.4 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t3 | 1 | 10 | 4.8 ms | 112 us | 2 us | 29.1 bits | 27.6 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t4 | 1 | 10 | 4.9 ms | 128 us | 4 us | 29.2 bits | 27.6 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t5 | 1 | 10 | 4.9 ms | 195 us | 9 us | 29.2 bits | 27.4 bits | -1.9 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t6 | 1 | 10 | 4.8 ms | 124 us | 18 us | 29.2 bits | 27.5 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t7 | 1 | 10 | 4.9 ms | 206 us | 37 us | 29.2 bits | 27.6 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t8 | 1 | 10 | 4.9 ms | 182 us | 75 us | 29.1 bits | 27.6 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t9 | 1 | 10 | 4.9 ms | 293 us | 154 us | 29.1 bits | 27.7 bits | -1.4 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t10 | 1 | 10 | 4.8 ms | 158 us | 302 us | 29.2 bits | 27.5 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | WH_log2t2 | 1 | 10 | 4.9 ms | 222 us | 1 us | 29.2 bits | 27.6 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | WH_log2t3 | 1 | 10 | 4.7 ms | 47 us | 2 us | 29.0 bits | 27.5 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | WH_log2t4 | 1 | 10 | 4.9 ms | 259 us | 4 us | 28.9 bits | 27.8 bits | -1.1 bits | yes | 0 |
| basic | 15 | native | WH_log2t5 | 1 | 10 | 4.8 ms | 106 us | 9 us | 29.2 bits | 27.9 bits | -1.3 bits | yes | 0 |
| basic | 15 | native | WH_log2t6 | 1 | 10 | 5.0 ms | 129 us | 19 us | 29.1 bits | 27.8 bits | -1.3 bits | yes | 0 |
| basic | 15 | native | WH_log2t7 | 1 | 10 | 4.9 ms | 150 us | 38 us | 29.2 bits | 27.6 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | WH_log2t8 | 1 | 10 | 4.9 ms | 204 us | 77 us | 29.2 bits | 27.4 bits | -1.8 bits | yes | 0 |
| basic | 15 | native | WH_log2t9 | 1 | 10 | 4.8 ms | 104 us | 150 us | 29.1 bits | 27.7 bits | -1.3 bits | yes | 0 |
| basic | 15 | native | WH_log2t10 | 1 | 10 | 4.8 ms | 109 us | 300 us | 29.2 bits | 27.7 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | IND_log2t2 | 1 | 10 | 4.8 ms | 119 us | 1 us | 29.2 bits | 27.5 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | IND_log2t3 | 1 | 10 | 4.9 ms | 131 us | 2 us | 29.3 bits | 27.4 bits | -1.9 bits | yes | 0 |
| basic | 15 | native | IND_log2t4 | 1 | 10 | 4.8 ms | 59 us | 4 us | 29.2 bits | 27.7 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | IND_log2t5 | 1 | 10 | 4.8 ms | 148 us | 9 us | 28.9 bits | 27.7 bits | -1.2 bits | yes | 0 |
| basic | 15 | native | IND_log2t6 | 1 | 10 | 4.8 ms | 105 us | 18 us | 29.2 bits | 27.9 bits | -1.3 bits | yes | 0 |
| basic | 15 | native | IND_log2t7 | 1 | 10 | 4.9 ms | 177 us | 38 us | 29.0 bits | 27.8 bits | -1.1 bits | yes | 0 |
| basic | 15 | native | IND_log2t8 | 1 | 10 | 4.8 ms | 155 us | 75 us | 29.2 bits | 27.8 bits | -1.4 bits | yes | 0 |
| basic | 15 | native | IND_log2t9 | 1 | 10 | 4.9 ms | 232 us | 152 us | 29.1 bits | 27.8 bits | -1.3 bits | yes | 0 |
| basic | 15 | native | IND_log2t10 | 1 | 10 | 4.9 ms | 166 us | 304 us | 29.2 bits | 27.6 bits | -1.6 bits | yes | 0 |

## Multi Thread Split Basic Native Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | native | BRU_log2t2 | 1 | 10 | 5.4 ms | 124 us | 0 us | 29.0 bits | 27.5 bits | -1.4 bits | yes | 0 |
| basic | 15 | native | BRU_log2t3 | 1 | 10 | 6.3 ms | 422 us | 0 us | 29.0 bits | 27.4 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | BRU_log2t4 | 1 | 10 | 10.1 ms | 1.0 ms | 1 us | 29.0 bits | 27.4 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | BRU_log2t5 | 1 | 10 | 20.7 ms | 1.2 ms | 1 us | 29.0 bits | 27.1 bits | -1.9 bits | yes | 0 |
| basic | 15 | native | BRU_log2t6 | 1 | 10 | 40.0 ms | 1.6 ms | 2 us | 28.9 bits | 27.4 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | BRU_log2t7 | 1 | 10 | 76.2 ms | 2.0 ms | 5 us | 28.9 bits | 27.2 bits | -1.8 bits | yes | 0 |
| basic | 15 | native | BRU_log2t8 | 1 | 10 | 149.5 ms | 5.0 ms | 9 us | 28.9 bits | 27.1 bits | -1.8 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t2 | 1 | 10 | 5.0 ms | 190 us | 0 us | 29.1 bits | 27.5 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t3 | 1 | 10 | 6.8 ms | 1.4 ms | 0 us | 29.0 bits | 27.5 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t4 | 1 | 10 | 8.9 ms | 1.4 ms | 1 us | 28.9 bits | 27.4 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t5 | 1 | 10 | 20.2 ms | 1.5 ms | 1 us | 28.9 bits | 27.5 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t6 | 1 | 10 | 37.2 ms | 784 us | 2 us | 28.9 bits | 27.2 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t7 | 1 | 10 | 75.3 ms | 2.1 ms | 5 us | 28.8 bits | 27.1 bits | -1.8 bits | yes | 0 |
| basic | 15 | native | LBRU_log2t8 | 1 | 10 | 146.7 ms | 2.1 ms | 9 us | 28.7 bits | 27.2 bits | -1.4 bits | yes | 0 |
| basic | 15 | native | WH_log2t2 | 1 | 10 | 5.5 ms | 846 us | 0 us | 29.1 bits | 27.6 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | WH_log2t3 | 1 | 10 | 7.1 ms | 1.3 ms | 0 us | 29.0 bits | 27.1 bits | -1.9 bits | yes | 0 |
| basic | 15 | native | WH_log2t4 | 1 | 10 | 10.1 ms | 1.0 ms | 1 us | 28.9 bits | 27.3 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | WH_log2t5 | 1 | 10 | 21.0 ms | 1.5 ms | 1 us | 28.9 bits | 27.4 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | WH_log2t6 | 1 | 10 | 38.4 ms | 956 us | 2 us | 28.9 bits | 27.3 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | WH_log2t7 | 1 | 10 | 74.1 ms | 2.0 ms | 5 us | 28.8 bits | 27.1 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | WH_log2t8 | 1 | 10 | 148.3 ms | 3.3 ms | 9 us | 28.8 bits | 27.1 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | IND_log2t2 | 1 | 10 | 5.2 ms | 217 us | 0 us | 29.1 bits | 27.6 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | IND_log2t3 | 1 | 10 | 7.5 ms | 1.8 ms | 0 us | 29.1 bits | 27.3 bits | -1.7 bits | yes | 0 |
| basic | 15 | native | IND_log2t4 | 1 | 10 | 11.1 ms | 1.3 ms | 1 us | 28.9 bits | 27.3 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | IND_log2t5 | 1 | 10 | 21.1 ms | 1.3 ms | 1 us | 28.9 bits | 27.5 bits | -1.5 bits | yes | 0 |
| basic | 15 | native | IND_log2t6 | 1 | 10 | 38.9 ms | 751 us | 2 us | 28.9 bits | 27.3 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | IND_log2t7 | 1 | 10 | 74.9 ms | 2.2 ms | 5 us | 28.9 bits | 27.4 bits | -1.6 bits | yes | 0 |
| basic | 15 | native | IND_log2t8 | 1 | 10 | 146.2 ms | 3.0 ms | 9 us | 28.8 bits | 27.1 bits | -1.6 bits | yes | 0 |

## Multi Thread Packed Basic Unary LUT Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | unary_lut | BRU_log2t2 | 1 | 10 | 8.7 ms | 37 us | 2 us | 29.3 bits | 27.4 bits | -1.9 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t3 | 1 | 10 | 16.2 ms | 427 us | 7 us | 29.0 bits | 27.6 bits | -1.4 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t4 | 1 | 10 | 24.4 ms | 552 us | 22 us | 29.2 bits | 27.6 bits | -1.7 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t5 | 1 | 10 | 33.8 ms | 538 us | 64 us | 29.1 bits | 27.3 bits | -1.8 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t6 | 1 | 10 | 54.3 ms | 644 us | 209 us | 29.2 bits | 27.4 bits | -1.7 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t7 | 1 | 10 | 78.4 ms | 387 us | 608 us | 29.1 bits | 27.5 bits | -1.6 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t8 | 1 | 10 | 123.3 ms | 1.7 ms | 1.9 ms | 29.1 bits | 27.4 bits | -1.6 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t9 | 1 | 10 | 209.0 ms | 3.1 ms | 6.5 ms | 29.0 bits | 27.2 bits | -1.8 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t10 | 1 | 10 | 341.2 ms | 1.1 ms | 21.3 ms | 29.1 bits | 26.9 bits | -2.1 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t2 | 1 | 10 | 7.8 ms | 252 us | 1 us | 29.2 bits | 27.6 bits | -1.6 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t3 | 1 | 10 | 15.6 ms | 235 us | 6 us | 29.2 bits | 27.2 bits | -2.0 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t4 | 1 | 10 | 20.4 ms | 83 us | 15 us | 29.2 bits | 27.1 bits | -2.1 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t5 | 1 | 10 | 33.5 ms | 317 us | 61 us | 29.1 bits | 27.1 bits | -2.0 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t6 | 1 | 10 | 50.3 ms | 753 us | 184 us | 28.9 bits | 27.4 bits | -1.5 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t7 | 1 | 10 | 78.3 ms | 639 us | 603 us | 29.2 bits | 27.1 bits | -2.1 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t8 | 1 | 10 | 124.7 ms | 745 us | 1.9 ms | 29.2 bits | 27.3 bits | -1.9 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t9 | 1 | 10 | 203.2 ms | 2.0 ms | 6.4 ms | 29.0 bits | 26.9 bits | -2.0 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t10 | 1 | 10 | 342.9 ms | 1.8 ms | 21.4 ms | 29.1 bits | 26.8 bits | -2.4 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t2 | 1 | 10 | 8.8 ms | 137 us | 2 us | 29.2 bits | 27.2 bits | -2.0 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t3 | 1 | 10 | 16.4 ms | 728 us | 7 us | 29.2 bits | 27.6 bits | -1.6 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t4 | 1 | 10 | 24.2 ms | 510 us | 22 us | 29.3 bits | 27.6 bits | -1.7 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t5 | 1 | 10 | 34.2 ms | 1.0 ms | 65 us | 29.1 bits | 27.5 bits | -1.6 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t6 | 1 | 10 | 54.0 ms | 424 us | 208 us | 29.1 bits | 27.6 bits | -1.5 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t7 | 1 | 10 | 79.5 ms | 844 us | 616 us | 29.1 bits | 27.4 bits | -1.7 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t8 | 1 | 10 | 123.5 ms | 1.1 ms | 1.9 ms | 29.1 bits | 27.4 bits | -1.7 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t9 | 1 | 10 | 201.8 ms | 585 us | 6.3 ms | 29.2 bits | 27.0 bits | -2.2 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t10 | 1 | 10 | 341.2 ms | 1.2 ms | 21.3 ms | 29.2 bits | 26.6 bits | -2.6 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t2 | 1 | 10 | 8.9 ms | 312 us | 2 us | 29.1 bits | 27.4 bits | -1.7 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t3 | 1 | 10 | 12.3 ms | 175 us | 5 us | 29.1 bits | 26.8 bits | -2.3 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t4 | 1 | 10 | 20.1 ms | 326 us | 18 us | 29.1 bits | 26.6 bits | -2.6 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t5 | 1 | 10 | 28.8 ms | 233 us | 55 us | 29.2 bits | 26.0 bits | -3.2 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t6 | 1 | 10 | 46.4 ms | 858 us | 178 us | 29.2 bits | 25.7 bits | -3.5 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t7 | 1 | 10 | 71.0 ms | 1.4 ms | 551 us | 29.0 bits | 25.1 bits | -3.9 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t8 | 1 | 10 | 107.7 ms | 812 us | 1.7 ms | 28.9 bits | 24.9 bits | -4.0 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t9 | 1 | 10 | 174.1 ms | 910 us | 5.4 ms | 28.9 bits | 24.2 bits | -4.8 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t10 | 1 | 10 | 276.5 ms | 2.0 ms | 17.3 ms | 29.2 bits | 24.0 bits | -5.1 bits | yes | 0 |

## Multi Thread Split Basic Unary LUT Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | unary_lut | BRU_log2t2 | 1 | 10 | 1.8 ms | 427 us | 0 us | 29.0 bits | 27.5 bits | -1.5 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t3 | 1 | 10 | 4.7 ms | 758 us | 0 us | 29.0 bits | 27.5 bits | -1.5 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t4 | 1 | 10 | 31.8 ms | 3.8 ms | 2 us | 29.0 bits | 27.3 bits | -1.7 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t5 | 1 | 10 | 109.4 ms | 5.2 ms | 7 us | 28.9 bits | 27.4 bits | -1.5 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t6 | 1 | 10 | 409.2 ms | 5.3 ms | 25 us | 28.8 bits | 27.3 bits | -1.5 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t7 | 1 | 10 | 1.69 s | 20.3 ms | 103 us | 28.9 bits | 27.3 bits | -1.6 bits | yes | 0 |
| basic | 15 | unary_lut | BRU_log2t8 | 1 | 10 | 7.11 s | 65.0 ms | 434 us | 28.9 bits | 27.0 bits | -1.9 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t2 | 1 | 10 | 1.1 ms | 114 us | 0 us | 29.1 bits | 29.1 bits | +0.0 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t3 | 1 | 10 | 3.4 ms | 496 us | 0 us | 29.0 bits | 27.5 bits | -1.5 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t4 | 1 | 10 | 19.6 ms | 2.3 ms | 1 us | 28.9 bits | 27.4 bits | -1.5 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t5 | 1 | 10 | 108.2 ms | 8.1 ms | 7 us | 29.0 bits | 27.0 bits | -2.0 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t6 | 1 | 10 | 380.5 ms | 3.0 ms | 23 us | 28.9 bits | 27.3 bits | -1.6 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t7 | 1 | 10 | 1.69 s | 15.7 ms | 103 us | 28.8 bits | 27.2 bits | -1.6 bits | yes | 0 |
| basic | 15 | unary_lut | LBRU_log2t8 | 1 | 10 | 6.72 s | 33.6 ms | 410 us | 28.9 bits | 27.0 bits | -1.8 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t2 | 1 | 10 | 1.7 ms | 322 us | 0 us | 29.1 bits | 27.4 bits | -1.6 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t3 | 1 | 10 | 3.9 ms | 532 us | 0 us | 29.0 bits | 27.5 bits | -1.5 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t4 | 1 | 10 | 26.8 ms | 3.6 ms | 2 us | 29.1 bits | 27.4 bits | -1.7 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t5 | 1 | 10 | 101.0 ms | 10.0 ms | 6 us | 29.0 bits | 27.5 bits | -1.5 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t6 | 1 | 10 | 380.0 ms | 6.8 ms | 23 us | 28.9 bits | 27.3 bits | -1.6 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t7 | 1 | 10 | 1.58 s | 10.1 ms | 97 us | 28.9 bits | 27.1 bits | -1.8 bits | yes | 0 |
| basic | 15 | unary_lut | WH_log2t8 | 1 | 10 | 6.70 s | 48.6 ms | 409 us | 28.8 bits | 27.3 bits | -1.5 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t2 | 1 | 10 | 1.8 ms | 260 us | 0 us | 29.1 bits | 28.4 bits | -0.7 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t3 | 1 | 10 | 2.6 ms | 324 us | 0 us | 29.0 bits | 27.7 bits | -1.3 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t4 | 1 | 10 | 5.0 ms | 620 us | 0 us | 28.9 bits | 27.2 bits | -1.7 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t5 | 1 | 10 | 10.4 ms | 891 us | 1 us | 28.9 bits | 26.8 bits | -2.1 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t6 | 1 | 10 | 20.2 ms | 1.8 ms | 1 us | 28.9 bits | 25.9 bits | -3.0 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t7 | 1 | 10 | 45.0 ms | 4.6 ms | 3 us | 28.9 bits | 25.7 bits | -3.2 bits | yes | 0 |
| basic | 15 | unary_lut | IND_log2t8 | 1 | 10 | 80.0 ms | 11.9 ms | 5 us | 28.8 bits | 25.2 bits | -3.6 bits | yes | 0 |

## Multi Thread Packed Basic Bivariate LUT Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | binary_lut | BRU_log2t2 | 3 | 10 | 74.7 ms | 871 us | 73 us | 29.2 bits | 27.3 bits | -1.9 bits | yes | 0 |
| basic | 15 | binary_lut | BRU_log2t3 | 3 | 10 | 134.9 ms | 1.6 ms | 527 us | 29.4 bits | 27.4 bits | -2.1 bits | yes | 0 |
| basic | 15 | binary_lut | BRU_log2t4 | 3 | 10 | 278.5 ms | 1.5 ms | 4.4 ms | 29.4 bits | 27.3 bits | -2.0 bits | yes | 0 |
| basic | 15 | binary_lut | BRU_log2t5 | 3 | 10 | 660.4 ms | 5.2 ms | 41.3 ms | 29.6 bits | 27.5 bits | -2.1 bits | yes | 0 |
| basic | 15 | binary_lut | LBRU_log2t2 | 3 | 10 | 50.3 ms | 1.2 ms | 28 us | 29.3 bits | 26.3 bits | -3.0 bits | yes | 0 |
| basic | 15 | binary_lut | LBRU_log2t3 | 3 | 10 | 114.2 ms | 3.2 ms | 342 us | 29.2 bits | 26.1 bits | -3.0 bits | yes | 0 |
| basic | 15 | binary_lut | LBRU_log2t4 | 3 | 10 | 220.3 ms | 2.6 ms | 2.3 ms | 29.5 bits | 25.8 bits | -3.7 bits | yes | 0 |
| basic | 15 | binary_lut | LBRU_log2t5 | 3 | 10 | 623.9 ms | 5.7 ms | 36.7 ms | 29.5 bits | 25.7 bits | -3.8 bits | yes | 0 |
| basic | 15 | binary_lut | WH_log2t2 | 3 | 10 | 75.7 ms | 1.3 ms | 74 us | 29.2 bits | 27.1 bits | -2.1 bits | yes | 0 |
| basic | 15 | binary_lut | WH_log2t3 | 3 | 10 | 135.8 ms | 1.9 ms | 531 us | 29.3 bits | 27.2 bits | -2.1 bits | yes | 0 |
| basic | 15 | binary_lut | WH_log2t4 | 3 | 10 | 282.0 ms | 3.7 ms | 4.4 ms | 29.4 bits | 27.3 bits | -2.1 bits | yes | 0 |
| basic | 15 | binary_lut | WH_log2t5 | 3 | 10 | 667.0 ms | 8.8 ms | 41.7 ms | 29.5 bits | 27.5 bits | -2.0 bits | yes | 0 |
| basic | 15 | binary_lut | IND_log2t2 | 3 | 10 | 71.2 ms | 939 us | 70 us | 29.3 bits | 25.9 bits | -3.3 bits | yes | 0 |
| basic | 15 | binary_lut | IND_log2t3 | 3 | 10 | 132.8 ms | 2.0 ms | 519 us | 29.1 bits | 24.9 bits | -4.1 bits | yes | 0 |
| basic | 15 | binary_lut | IND_log2t4 | 3 | 10 | 277.7 ms | 3.1 ms | 4.3 ms | 29.4 bits | 24.3 bits | -5.1 bits | yes | 0 |
| basic | 15 | binary_lut | IND_log2t5 | 3 | 10 | 662.0 ms | 5.0 ms | 41.4 ms | 29.3 bits | 23.6 bits | -5.6 bits | yes | 0 |

## Multi Thread Split Basic Bivariate LUT Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | binary_lut | BRU_log2t2 | 2 | 10 | 17.6 ms | 1.9 ms | 1 us | 29.2 bits | 19.1 bits | -10.1 bits | yes | 0 |
| basic | 15 | binary_lut | BRU_log2t3 | 2 | 10 | 125.6 ms | 9.1 ms | 8 us | 29.0 bits | 19.3 bits | -9.7 bits | yes | 0 |
| basic | 15 | binary_lut | BRU_log2t4 | 2 | 10 | 681.7 ms | 9.1 ms | 42 us | 29.0 bits | 19.5 bits | -9.5 bits | yes | 0 |
| basic | 15 | binary_lut | BRU_log2t5 | 2 | 10 | 4.60 s | 36.8 ms | 281 us | 28.8 bits | 19.7 bits | -9.1 bits | yes | 0 |
| basic | 15 | binary_lut | LBRU_log2t2 | 2 | 10 | 11.1 ms | 470 us | 1 us | 29.1 bits | 18.4 bits | -10.7 bits | yes | 0 |
| basic | 15 | binary_lut | LBRU_log2t3 | 2 | 10 | 85.1 ms | 7.8 ms | 5 us | 29.0 bits | 17.6 bits | -11.5 bits | yes | 0 |
| basic | 15 | binary_lut | LBRU_log2t4 | 2 | 10 | 393.2 ms | 4.2 ms | 24 us | 29.0 bits | 17.4 bits | -11.6 bits | yes | 0 |
| basic | 15 | binary_lut | LBRU_log2t5 | 2 | 10 | 4.15 s | 36.8 ms | 253 us | 28.9 bits | 17.4 bits | -11.5 bits | yes | 0 |
| basic | 15 | binary_lut | WH_log2t2 | 2 | 10 | 18.0 ms | 2.2 ms | 1 us | 28.9 bits | 18.6 bits | -10.3 bits | yes | 0 |
| basic | 15 | binary_lut | WH_log2t3 | 2 | 10 | 116.5 ms | 9.5 ms | 7 us | 29.0 bits | 19.1 bits | -9.9 bits | yes | 0 |
| basic | 15 | binary_lut | WH_log2t4 | 2 | 10 | 661.1 ms | 9.8 ms | 40 us | 29.0 bits | 19.0 bits | -9.9 bits | yes | 0 |
| basic | 15 | binary_lut | WH_log2t5 | 2 | 10 | 4.48 s | 34.6 ms | 273 us | 29.0 bits | 19.3 bits | -9.6 bits | yes | 0 |
| basic | 15 | binary_lut | IND_log2t2 | 2 | 10 | 17.0 ms | 1.6 ms | 1 us | 28.9 bits | 18.4 bits | -10.5 bits | yes | 0 |
| basic | 15 | binary_lut | IND_log2t3 | 2 | 10 | 83.8 ms | 4.3 ms | 5 us | 29.1 bits | 18.4 bits | -10.7 bits | yes | 0 |
| basic | 15 | binary_lut | IND_log2t4 | 2 | 10 | 380.8 ms | 15.6 ms | 23 us | 29.1 bits | 18.4 bits | -10.7 bits | yes | 0 |
| basic | 15 | binary_lut | IND_log2t5 | 2 | 10 | 1.48 s | 14.7 ms | 91 us | 28.9 bits | 18.4 bits | -10.6 bits | yes | 0 |

## Multi Thread Basic 4-Variate LUT Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | four_lut | BRU_log2t2 | 4 | 10 | 396.9 ms | 4.5 ms | 6.2 ms | 29.5 bits | 27.1 bits | -2.5 bits | yes | 0 |
| basic | 15 | four_lut | BRU_log2t3 | 4 | 10 | 2.45 s | 24.4 ms | 612.1 ms | 30.2 bits | 27.6 bits | -2.6 bits | yes | 0 |
| basic | 15 | four_lut | LBRU_log2t2 | 4 | 10 | 217.3 ms | 1.6 ms | 1.1 ms | 29.7 bits | 24.8 bits | -4.9 bits | yes | 0 |
| basic | 15 | four_lut | LBRU_log2t3 | 4 | 10 | 1.68 s | 53.7 ms | 280.4 ms | 29.9 bits | 24.9 bits | -5.0 bits | yes | 0 |
| basic | 15 | four_lut | WH_log2t2 | 4 | 10 | 402.4 ms | 4.1 ms | 6.3 ms | 29.4 bits | 27.2 bits | -2.3 bits | yes | 0 |
| basic | 15 | four_lut | WH_log2t3 | 4 | 10 | 2.54 s | 23.3 ms | 634.5 ms | 29.5 bits | 27.2 bits | -2.4 bits | yes | 0 |
| basic | 15 | four_lut | IND_log2t2 | 4 | 10 | 398.6 ms | 6.7 ms | 6.2 ms | 29.4 bits | 23.9 bits | -5.5 bits | yes | 0 |
| basic | 15 | four_lut | IND_log2t3 | 4 | 10 | 2.51 s | 11.5 ms | 628.4 ms | 29.6 bits | 22.6 bits | -7.1 bits | yes | 0 |

## Multi Thread Packed Basic Clean Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | clean | BRU_log2t2 | 4 | 10 | 57.7 ms | 354 us | 11 us | 11.8 bits | 16.5 bits | +4.7 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t3 | 4 | 10 | 89.7 ms | 312 us | 38 us | 11.8 bits | 16.5 bits | +4.7 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t4 | 4 | 10 | 126.0 ms | 374 us | 115 us | 11.8 bits | 16.5 bits | +4.7 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t5 | 4 | 10 | 166.7 ms | 617 us | 316 us | 11.8 bits | 16.5 bits | +4.7 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t6 | 4 | 10 | 248.8 ms | 1.3 ms | 957 us | 11.8 bits | 16.5 bits | +4.7 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t7 | 4 | 10 | 345.2 ms | 2.1 ms | 2.7 ms | 11.8 bits | 16.5 bits | +4.7 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t8 | 4 | 10 | 517.9 ms | 1.2 ms | 8.1 ms | 11.8 bits | 16.5 bits | +4.7 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t9 | 4 | 10 | 803.3 ms | 4.8 ms | 25.1 ms | 11.8 bits | 16.5 bits | +4.7 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t10 | 4 | 10 | 1.29 s | 12.2 ms | 80.7 ms | 11.8 bits | 16.5 bits | +4.6 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t2 | 4 | 10 | 54.7 ms | 746 us | 7 us | 11.8 bits | 17.8 bits | +6.0 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t3 | 4 | 10 | 89.0 ms | 674 us | 33 us | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t4 | 4 | 10 | 109.8 ms | 951 us | 80 us | 11.8 bits | 17.8 bits | +6.0 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t5 | 4 | 10 | 165.6 ms | 1.4 ms | 303 us | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t6 | 4 | 10 | 231.9 ms | 2.0 ms | 849 us | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t7 | 4 | 10 | 344.6 ms | 3.6 ms | 2.7 ms | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t8 | 4 | 10 | 519.6 ms | 2.0 ms | 8.0 ms | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t9 | 4 | 10 | 799.0 ms | 2.1 ms | 25.0 ms | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t10 | 4 | 10 | 1.30 s | 3.1 ms | 81.3 ms | 11.8 bits | 17.7 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | WH_log2t2 | 2 | 10 | 12.9 ms | 337 us | 2 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t3 | 2 | 10 | 13.0 ms | 449 us | 6 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t4 | 2 | 10 | 12.8 ms | 271 us | 12 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t5 | 2 | 10 | 12.8 ms | 288 us | 24 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t6 | 2 | 10 | 12.9 ms | 256 us | 49 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t7 | 2 | 10 | 12.8 ms | 354 us | 99 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t8 | 2 | 10 | 12.9 ms | 442 us | 201 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t9 | 2 | 10 | 12.7 ms | 239 us | 398 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t10 | 2 | 10 | 12.7 ms | 275 us | 797 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | IND_log2t2 | 2 | 10 | 12.9 ms | 262 us | 2 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t3 | 2 | 10 | 12.8 ms | 329 us | 5 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t4 | 2 | 10 | 13.0 ms | 336 us | 12 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t5 | 2 | 10 | 13.1 ms | 350 us | 25 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t6 | 2 | 10 | 12.9 ms | 359 us | 49 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t7 | 2 | 10 | 12.9 ms | 380 us | 100 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t8 | 2 | 10 | 12.9 ms | 304 us | 201 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t9 | 2 | 10 | 12.9 ms | 284 us | 403 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t10 | 2 | 10 | 12.7 ms | 225 us | 792 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |

## Multi Thread Split Basic Clean Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | clean | BRU_log2t2 | 4 | 10 | 27.0 ms | 2.4 ms | 2 us | 11.8 bits | 16.8 bits | +5.0 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t3 | 4 | 10 | 42.7 ms | 3.0 ms | 3 us | 11.8 bits | 16.8 bits | +5.0 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t4 | 4 | 10 | 158.0 ms | 11.9 ms | 10 us | 11.8 bits | 16.8 bits | +5.0 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t5 | 4 | 10 | 530.1 ms | 9.3 ms | 32 us | 11.8 bits | 16.8 bits | +5.0 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t6 | 4 | 10 | 1.97 s | 11.8 ms | 120 us | 11.8 bits | 16.8 bits | +5.0 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t7 | 4 | 10 | 7.65 s | 45.0 ms | 467 us | 11.8 bits | 16.8 bits | +4.9 bits | yes | 0 |
| basic | 15 | clean | BRU_log2t8 | 4 | 10 | 30.40 s | 100.9 ms | 1.9 ms | 11.8 bits | 16.8 bits | +4.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t2 | 4 | 10 | 25.0 ms | 1.3 ms | 2 us | 11.8 bits | 17.8 bits | +6.0 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t3 | 4 | 10 | 39.3 ms | 2.4 ms | 2 us | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t4 | 4 | 10 | 106.9 ms | 6.7 ms | 7 us | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t5 | 4 | 10 | 501.4 ms | 5.7 ms | 31 us | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t6 | 4 | 10 | 1.81 s | 7.8 ms | 111 us | 11.8 bits | 17.8 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t7 | 4 | 10 | 7.49 s | 11.7 ms | 457 us | 11.8 bits | 17.7 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | LBRU_log2t8 | 4 | 10 | 29.26 s | 141.0 ms | 1.8 ms | 11.8 bits | 17.7 bits | +5.9 bits | yes | 0 |
| basic | 15 | clean | WH_log2t2 | 2 | 10 | 14.2 ms | 326 us | 1 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t3 | 2 | 10 | 17.3 ms | 991 us | 1 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t4 | 2 | 10 | 28.7 ms | 1.7 ms | 2 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t5 | 2 | 10 | 69.5 ms | 2.2 ms | 4 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t6 | 2 | 10 | 139.4 ms | 6.3 ms | 9 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t7 | 2 | 10 | 265.6 ms | 8.2 ms | 16 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | WH_log2t8 | 2 | 10 | 537.5 ms | 9.0 ms | 33 us | 11.8 bits | 21.1 bits | +9.3 bits | yes | 0 |
| basic | 15 | clean | IND_log2t2 | 2 | 10 | 14.0 ms | 298 us | 1 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t3 | 2 | 10 | 17.4 ms | 1.9 ms | 1 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t4 | 2 | 10 | 28.2 ms | 2.2 ms | 2 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t5 | 2 | 10 | 66.9 ms | 3.0 ms | 4 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t6 | 2 | 10 | 139.2 ms | 4.4 ms | 8 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t7 | 2 | 10 | 266.3 ms | 8.2 ms | 16 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |
| basic | 15 | clean | IND_log2t8 | 2 | 10 | 537.5 ms | 9.1 ms | 33 us | 11.8 bits | 18.7 bits | +6.8 bits | yes | 0 |

## Multi Thread Basic Split->Standard Results

| Target | LogN | Operation | Shape | Levels | Remaining | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | --------: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | to_standard | BRU_log2t2 | 1 | 0 | 10 | 1.6 ms | 181 us | 0 us |  |  |  | yes | 0 |
| basic | 15 | to_standard | BRU_log2t3 | 1 | 0 | 10 | 2.6 ms | 190 us | 0 us |  |  |  | yes | 0 |
| basic | 15 | to_standard | BRU_log2t4 | 1 | 0 | 10 | 4.0 ms | 98 us | 0 us |  |  |  | yes | 0 |
| basic | 15 | to_standard | BRU_log2t5 | 1 | 0 | 10 | 7.4 ms | 257 us | 0 us |  |  |  | yes | 0 |
| basic | 15 | to_standard | BRU_log2t6 | 1 | 0 | 10 | 14.1 ms | 367 us | 1 us |  |  |  | yes | 0 |
| basic | 15 | to_standard | BRU_log2t7 | 1 | 0 | 10 | 27.4 ms | 531 us | 2 us |  |  |  | yes | 0 |
| basic | 15 | to_standard | BRU_log2t8 | 1 | 0 | 10 | 54.1 ms | 1.8 ms | 3 us |  |  |  | yes | 0 |

## Multi Thread Basic Standard->Split Results

| Target | LogN | Operation | Shape | Levels | Remaining | Samples | Mean | Stddev | Amortized | Input Precision | Output Precision | Gain/Loss | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | --------: | ------: | ---: | -----: | --------: | --------------: | ---------------: | --------: | :------ | --------: |
| basic | 15 | from_standard | BRU_log2t2 | 9 | 6 | 10 | 3.78 s | 75.6 ms | 231 us |  |  |  | yes | 0 |
| basic | 15 | from_standard | BRU_log2t3 | 9 | 6 | 10 | 4.15 s | 30.0 ms | 253 us |  |  |  | yes | 0 |
| basic | 15 | from_standard | BRU_log2t4 | 9 | 6 | 10 | 5.34 s | 32.5 ms | 326 us |  |  |  | yes | 0 |
| basic | 15 | from_standard | BRU_log2t5 | 10 | 5 | 10 | 5.85 s | 75.6 ms | 357 us |  |  |  | yes | 0 |
| basic | 15 | from_standard | BRU_log2t6 | 12 | 3 | 10 | 6.28 s | 24.3 ms | 383 us |  |  |  | yes | 0 |
| basic | 15 | from_standard | BRU_log2t7 | 13 | 2 | 10 | 6.96 s | 87.1 ms | 425 us |  |  |  | yes | 0 |

## Multi Thread CRT Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | :------ | --------: |
| crt | 15 | add | 64bits | 1 | 10 | 5.2 ms | 960 us | 119 us | yes | 0 |
| crt | 15 | add | 256bits | 1 | 10 | 5.0 ms | 193 us | 1.2 ms | yes | 0 |
| crt | 15 | sub | 64bits | 1 | 10 | 8.6 ms | 161 us | 196 us | yes | 0 |
| crt | 15 | sub | 256bits | 1 | 10 | 8.6 ms | 195 us | 2.1 ms | yes | 0 |
| crt | 15 | mul_lbru | 64bits | 1 | 10 | 4.9 ms | 121 us | 111 us | yes | 0 |
| crt | 15 | mul_lbru | 256bits | 1 | 10 | 5.0 ms | 170 us | 1.2 ms | yes | 0 |
| crt | 15 | bru_to_lbru | 64bits | 1 | 10 | 48.9 ms | 561 us | 1.1 ms | yes | 0 |
| crt | 15 | bru_to_lbru | 256bits | 1 | 10 | 105.1 ms | 914 us | 26.3 ms | yes | 0 |
| crt | 15 | lbru_to_bru | 64bits | 1 | 10 | 48.5 ms | 286 us | 1.1 ms | yes | 0 |
| crt | 15 | lbru_to_bru | 256bits | 1 | 10 | 107.2 ms | 2.0 ms | 26.8 ms | yes | 0 |
| crt | 15 | clean | 64bits | 4 | 10 | 225.9 ms | 1.7 ms | 5.1 ms | yes | 0 |
| crt | 15 | clean | 256bits | 4 | 10 | 443.8 ms | 1.3 ms | 111.0 ms | yes | 0 |

## Multi Thread Radix Results

| Target | LogN | Operation | Shape | Levels | Samples | Mean | Stddev | Amortized | Correct | Max Error |
| :----- | ---: | :-------- | :---- | -----: | ------: | ---: | -----: | --------: | :------ | --------: |
| radix | 16 | Add | 64bit_r4 | 9 | 10 | 1.55 s | 8.9 ms | 36.9 ms | yes | 0 |
| radix | 16 | Add | 256bit_r4 | 11 | 10 | 2.34 s | 19.9 ms | 234.4 ms | yes | 0 |
| radix | 16 | Sub | 64bit_r4 | 9 | 10 | 1.56 s | 17.2 ms | 37.1 ms | yes | 0 |
| radix | 16 | Sub | 256bit_r4 | 11 | 10 | 2.34 s | 10.3 ms | 234.4 ms | yes | 0 |
| radix | 16 | Eq | 64bit_r4 | 8 | 10 | 831.1 ms | 5.6 ms | 19.8 ms | yes | 0 |
| radix | 16 | Eq | 256bit_r4 | 10 | 10 | 1.29 s | 6.7 ms | 129.3 ms | yes | 0 |
| radix | 16 | Lt | 64bit_r4 | 8 | 10 | 1.03 s | 8.6 ms | 24.6 ms | yes | 0 |
| radix | 16 | Lt | 256bit_r4 | 10 | 10 | 1.67 s | 6.4 ms | 166.7 ms | yes | 0 |
| radix | 16 | Cmp | 64bit_r4 | 8 | 10 | 1.08 s | 8.5 ms | 25.8 ms | yes | 0 |
| radix | 16 | Cmp | 256bit_r4 | 10 | 10 | 1.76 s | 10.4 ms | 176.2 ms | yes | 0 |
| radix | 16 | Clean | 64bit_r4 | 4 | 10 | 120.3 ms | 677 us | 2.9 ms | yes | 0 |
| radix | 16 | Clean | 256bit_r4 | 4 | 10 | 120.0 ms | 556 us | 12.0 ms | yes | 0 |


## Basic Native Parameter Check

| Target | LogN | Operation | LogQ | LogP | LogQP | Max LogQP at h=256 | Secure |
| :----- | ---: | :-------- | :--- | :--- | ----: | ------------------: | :----- |
| basic | 15 | basic native BRU_log2t2 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native BRU_log2t3 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native BRU_log2t4 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native BRU_log2t5 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native BRU_log2t6 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native BRU_log2t7 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native BRU_log2t8 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native BRU_log2t9 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native BRU_log2t10 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native LBRU_log2t2 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native LBRU_log2t3 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native LBRU_log2t4 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native LBRU_log2t5 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native LBRU_log2t6 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native LBRU_log2t7 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native LBRU_log2t8 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native LBRU_log2t9 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native LBRU_log2t10 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native WH_log2t2 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native WH_log2t3 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native WH_log2t4 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native WH_log2t5 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native WH_log2t6 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native WH_log2t7 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native WH_log2t8 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native WH_log2t9 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native WH_log2t10 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native IND_log2t2 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native IND_log2t3 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native IND_log2t4 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native IND_log2t5 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native IND_log2t6 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native IND_log2t7 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native IND_log2t8 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native IND_log2t9 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic native IND_log2t10 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |

## Basic Unary LUT Parameter Check

| Target | LogN | Operation | LogQ | LogP | LogQP | Max LogQP at h=256 | Secure |
| :----- | ---: | :-------- | :--- | :--- | ----: | ------------------: | :----- |
| basic | 15 | basic unary_lut BRU_log2t2 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut BRU_log2t3 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut BRU_log2t4 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut BRU_log2t5 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut BRU_log2t6 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut BRU_log2t7 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut BRU_log2t8 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut BRU_log2t9 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut BRU_log2t10 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut LBRU_log2t2 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut LBRU_log2t3 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut LBRU_log2t4 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut LBRU_log2t5 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut LBRU_log2t6 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut LBRU_log2t7 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut LBRU_log2t8 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut LBRU_log2t9 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut LBRU_log2t10 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut WH_log2t2 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut WH_log2t3 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut WH_log2t4 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut WH_log2t5 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut WH_log2t6 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut WH_log2t7 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut WH_log2t8 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut WH_log2t9 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut WH_log2t10 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut IND_log2t2 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut IND_log2t3 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut IND_log2t4 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut IND_log2t5 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut IND_log2t6 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut IND_log2t7 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut IND_log2t8 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut IND_log2t9 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic unary_lut IND_log2t10 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |

## Basic Bivariate LUT Parameter Check

| Target | LogN | Operation | LogQ | LogP | LogQP | Max LogQP at h=256 | Secure |
| :----- | ---: | :-------- | :--- | :--- | ----: | ------------------: | :----- |
| basic | 15 | basic binary_lut BRU_log2t2 | `[55, 40 x 3]` | `[60]` | 235.0 | 774 | yes |
| basic | 15 | basic binary_lut BRU_log2t3 | `[55, 40 x 3]` | `[60]` | 235.0 | 774 | yes |
| basic | 15 | basic binary_lut BRU_log2t4 | `[55, 40 x 3]` | `[60]` | 235.0 | 774 | yes |
| basic | 15 | basic binary_lut BRU_log2t5 | `[55, 40 x 3]` | `[60]` | 235.0 | 774 | yes |
| basic | 15 | basic binary_lut LBRU_log2t2 | `[55, 40 x 3]` | `[60]` | 235.0 | 774 | yes |
| basic | 15 | basic binary_lut LBRU_log2t3 | `[55, 40 x 3]` | `[60]` | 235.0 | 774 | yes |
| basic | 15 | basic binary_lut LBRU_log2t4 | `[55, 40 x 3]` | `[60]` | 235.0 | 774 | yes |
| basic | 15 | basic binary_lut LBRU_log2t5 | `[55, 40 x 3]` | `[60]` | 235.0 | 774 | yes |
| basic | 15 | basic binary_lut WH_log2t2 | `[55, 40 x 3]` | `[60]` | 235.0 | 774 | yes |
| basic | 15 | basic binary_lut WH_log2t3 | `[55, 40 x 3]` | `[60]` | 235.0 | 774 | yes |
| basic | 15 | basic binary_lut WH_log2t4 | `[55, 40 x 3]` | `[60]` | 235.0 | 774 | yes |
| basic | 15 | basic binary_lut WH_log2t5 | `[55, 40 x 3]` | `[60]` | 235.0 | 774 | yes |
| basic | 15 | basic binary_lut IND_log2t2 | `[55, 40 x 3]` | `[60]` | 235.0 | 774 | yes |
| basic | 15 | basic binary_lut IND_log2t3 | `[55, 40 x 3]` | `[60]` | 235.0 | 774 | yes |
| basic | 15 | basic binary_lut IND_log2t4 | `[55, 40 x 3]` | `[60]` | 235.0 | 774 | yes |
| basic | 15 | basic binary_lut IND_log2t5 | `[55, 40 x 3]` | `[60]` | 235.0 | 774 | yes |

## Basic 4-Variate LUT Parameter Check

| Target | LogN | Operation | LogQ | LogP | LogQP | Max LogQP at h=256 | Secure |
| :----- | ---: | :-------- | :--- | :--- | ----: | ------------------: | :----- |
| basic | 15 | basic four_lut BRU_log2t2 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic four_lut BRU_log2t3 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic four_lut LBRU_log2t2 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic four_lut LBRU_log2t3 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic four_lut WH_log2t2 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic four_lut WH_log2t3 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic four_lut IND_log2t2 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic four_lut IND_log2t3 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |

## Basic Clean Parameter Check

| Target | LogN | Operation | LogQ | LogP | LogQP | Max LogQP at h=256 | Secure |
| :----- | ---: | :-------- | :--- | :--- | ----: | ------------------: | :----- |
| basic | 15 | basic clean BRU_log2t2 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic clean BRU_log2t3 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic clean BRU_log2t4 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic clean BRU_log2t5 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic clean BRU_log2t6 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic clean BRU_log2t7 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic clean BRU_log2t8 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic clean BRU_log2t9 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic clean BRU_log2t10 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic clean LBRU_log2t2 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic clean LBRU_log2t3 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic clean LBRU_log2t4 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic clean LBRU_log2t5 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic clean LBRU_log2t6 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic clean LBRU_log2t7 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic clean LBRU_log2t8 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic clean LBRU_log2t9 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic clean LBRU_log2t10 | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |
| basic | 15 | basic clean WH_log2t2 | `[55, 40 x 2]` | `[60]` | 195.0 | 774 | yes |
| basic | 15 | basic clean WH_log2t3 | `[55, 40 x 2]` | `[60]` | 195.0 | 774 | yes |
| basic | 15 | basic clean WH_log2t4 | `[55, 40 x 2]` | `[60]` | 195.0 | 774 | yes |
| basic | 15 | basic clean WH_log2t5 | `[55, 40 x 2]` | `[60]` | 195.0 | 774 | yes |
| basic | 15 | basic clean WH_log2t6 | `[55, 40 x 2]` | `[60]` | 195.0 | 774 | yes |
| basic | 15 | basic clean WH_log2t7 | `[55, 40 x 2]` | `[60]` | 195.0 | 774 | yes |
| basic | 15 | basic clean WH_log2t8 | `[55, 40 x 2]` | `[60]` | 195.0 | 774 | yes |
| basic | 15 | basic clean WH_log2t9 | `[55, 40 x 2]` | `[60]` | 195.0 | 774 | yes |
| basic | 15 | basic clean WH_log2t10 | `[55, 40 x 2]` | `[60]` | 195.0 | 774 | yes |
| basic | 15 | basic clean IND_log2t2 | `[55, 40 x 2]` | `[60]` | 195.0 | 774 | yes |
| basic | 15 | basic clean IND_log2t3 | `[55, 40 x 2]` | `[60]` | 195.0 | 774 | yes |
| basic | 15 | basic clean IND_log2t4 | `[55, 40 x 2]` | `[60]` | 195.0 | 774 | yes |
| basic | 15 | basic clean IND_log2t5 | `[55, 40 x 2]` | `[60]` | 195.0 | 774 | yes |
| basic | 15 | basic clean IND_log2t6 | `[55, 40 x 2]` | `[60]` | 195.0 | 774 | yes |
| basic | 15 | basic clean IND_log2t7 | `[55, 40 x 2]` | `[60]` | 195.0 | 774 | yes |
| basic | 15 | basic clean IND_log2t8 | `[55, 40 x 2]` | `[60]` | 195.0 | 774 | yes |
| basic | 15 | basic clean IND_log2t9 | `[55, 40 x 2]` | `[60]` | 195.0 | 774 | yes |
| basic | 15 | basic clean IND_log2t10 | `[55, 40 x 2]` | `[60]` | 195.0 | 774 | yes |

## Basic Split->Standard Parameter Check

| Target | LogN | Operation | LogQ | LogP | LogQP | Max LogQP at h=256 | Secure |
| :----- | ---: | :-------- | :--- | :--- | ----: | ------------------: | :----- |
| basic | 15 | basic to_standard BRU_log2t2 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic to_standard BRU_log2t3 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic to_standard BRU_log2t4 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic to_standard BRU_log2t5 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic to_standard BRU_log2t6 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic to_standard BRU_log2t7 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| basic | 15 | basic to_standard BRU_log2t8 | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |

## Basic Standard->Split Parameter Check

| Target | LogN | Operation | LogQ | LogP | LogQP | Max LogQP at h=256 | Secure |
| :----- | ---: | :-------- | :--- | :--- | ----: | ------------------: | :----- |
| basic | 15 | basic from_standard BRU_log2t2 | `[55, 40 x 9]` | `[60]` | 770.0 | 774 | yes |
| basic | 15 | basic from_standard BRU_log2t3 | `[55, 40 x 9]` | `[60]` | 770.0 | 774 | yes |
| basic | 15 | basic from_standard BRU_log2t4 | `[55, 40 x 9]` | `[60]` | 770.0 | 774 | yes |
| basic | 15 | basic from_standard BRU_log2t5 | `[55, 40 x 10]` | `[60]` | 770.0 | 774 | yes |
| basic | 15 | basic from_standard BRU_log2t6 | `[55, 40 x 12]` | `[60]` | 770.0 | 774 | yes |
| basic | 15 | basic from_standard BRU_log2t7 | `[55, 40 x 13]` | `[60]` | 770.0 | 774 | yes |

## CRT Parameter Check

| Target | LogN | Operation | LogQ | LogP | LogQP | Max LogQP at h=256 | Secure |
| :----- | ---: | :-------- | :--- | :--- | ----: | ------------------: | :----- |
| crt | 15 | CRT add | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| crt | 15 | CRT sub | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| crt | 15 | CRT mul_lbru | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| crt | 15 | CRT bru_to_lbru | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| crt | 15 | CRT lbru_to_bru | `[55, 40 x 1]` | `[60]` | 155.0 | 774 | yes |
| crt | 15 | CRT clean | `[55, 40 x 4]` | `[60]` | 275.0 | 774 | yes |

## Radix Parameter Check

| Target | LogN | Operation | LogQ | LogP | LogQP | Max LogQP at h=256 | Secure |
| :----- | ---: | :-------- | :--- | :--- | ----: | ------------------: | :----- |
| radix | 16 | Add 64bit_r4 | `[60, 40 x 9]` | `[60]` | 480.0 | 1553 | yes |
| radix | 16 | Add 256bit_r4 | `[60, 40 x 11]` | `[60]` | 560.0 | 1553 | yes |
| radix | 16 | Sub 64bit_r4 | `[60, 40 x 9]` | `[60]` | 480.0 | 1553 | yes |
| radix | 16 | Sub 256bit_r4 | `[60, 40 x 11]` | `[60]` | 560.0 | 1553 | yes |
| radix | 16 | Eq 64bit_r4 | `[60, 40 x 8]` | `[60]` | 440.0 | 1553 | yes |
| radix | 16 | Eq 256bit_r4 | `[60, 40 x 10]` | `[60]` | 520.0 | 1553 | yes |
| radix | 16 | Lt 64bit_r4 | `[60, 40 x 8]` | `[60]` | 440.0 | 1553 | yes |
| radix | 16 | Lt 256bit_r4 | `[60, 40 x 10]` | `[60]` | 520.0 | 1553 | yes |
| radix | 16 | Cmp 64bit_r4 | `[60, 40 x 8]` | `[60]` | 440.0 | 1553 | yes |
| radix | 16 | Cmp 256bit_r4 | `[60, 40 x 10]` | `[60]` | 520.0 | 1553 | yes |
| radix | 16 | Clean 64bit_r4 | `[60, 40 x 4]` | `[60]` | 280.0 | 1553 | yes |
| radix | 16 | Clean 256bit_r4 | `[60, 40 x 4]` | `[60]` | 280.0 | 1553 | yes |
