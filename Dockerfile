FROM golang:1.23-bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends python3 ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
COPY . .

ENV GOCACHE=/tmp/go-build-character-encoding

ENTRYPOINT ["python3", "scripts/run_benchmarks.py"]
CMD ["--target", "radix", "--mode", "single", "--ops", "Add", "--shapes", "64bit_r4", "--count", "1"]
