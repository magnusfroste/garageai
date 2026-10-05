# Contributing to GarageAI

Thanks for helping build Europe's distributed AI capacity.

## Ways to help

- **Report a bug or suggest an idea** — open an [issue](https://github.com/magnusfroste/garageai/issues).
- **Improve the connect script** — [`scripts/garageai-connect.sh`](../scripts/garageai-connect.sh) runs on every garage; support for more runtimes and platforms is especially welcome.
- **Improve the gateway setup** — [`infra/gateway/`](../infra/gateway/README.md).
- **Improve the site** — content lives in [`src/content/`](../src/content).

## Workflow

```bash
git clone https://github.com/magnusfroste/garageai.git
cd garageai
npm install
npm run dev
```

1. Fork the repo and create a branch.
2. Keep each pull request to one change, and describe what it does and how you checked it.
3. Run `npm run lint` and `npm run build` before you open the pull request.

## License

By contributing you agree that your contributions are licensed under the [MIT License](../LICENSE).
