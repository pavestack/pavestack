# Pavestack Portal

Read-only developer portal for the Pavestack IDP. Displays service catalog entries and scorecards sourced from `catalog-info.yaml` and `scorecard.yaml` in each service directory.

## Development

```bash
npm ci
npm run dev
```

## Static export

```bash
npm run build
```

Output is written to `out/` and can be hosted on S3/CloudFront or any static file host.

## Data model

The build script scans:

- `service-template-api/`
- `services/*/`

It never mutates cluster or Git state.
