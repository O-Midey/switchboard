# Contributing

Choose work from [ROADMAP.md](ROADMAP.md) and carry its stable milestone ID into the issue and pull-request title. Open an issue before changing a public interface or architecture decision. Keep changes focused, add behavior-level tests, and run `make check` before opening a pull request. New routing adapters must be concurrency-safe, deterministic for the same state, and unable to select targets outside the supplied healthy set.

Never commit secrets, generated dashboard output, local environment files, or benchmark claims without reproducible evidence. Use conventional, imperative commit subjects; do not add AI attribution trailers.
