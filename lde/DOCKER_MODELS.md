# Docker Models Integration

This LDE deployment uses Docker's built-in AI models feature for ML capabilities.

## Models Configuration

The following models are configured in `docker-compose.yml`:

### 1. Text Embeddings Model
- **Name**: `embeddings`
- **Model**: `ai/embeddinggemma:latest`
- **Parameters**: 308M (Google DeepMind)
- **Context Size**: 2048 tokens
- **Use Case**: Convert text to vector embeddings for semantic search and RAG
- **Features**: Trained on 100+ languages, runs on <200MB RAM with quantization

### 2. Text Summarization Model
- **Name**: `summarize`
- **Model**: `ai/gpt-oss:latest`
- **Parameters**: 21B total (~3.6B active per token)
- **Architecture**: Sparse MoE (32 experts, 4 active per token)
- **Context Size**: 8192 tokens
- **Runtime Flags**:
  - `--temperature=0.7` (balanced creativity)
  - `--max-tokens=2048` (max output length)
- **Use Case**: Generate summaries, text completions, reasoning tasks
- **License**: Apache 2.0

## How It Works

1. **Model Declaration**: Models are declared in the `models:` section of `docker-compose.yml`
2. **Service Binding**: The Hugr service binds to these models via the `models:` directive
3. **Environment Variables**: Docker automatically injected environment variables into Hugr:
   - `EMBEDDINGS_URL` - Endpoint for embeddings model
   - `EMBEDDINGS_MODEL` - Model identifier
   - `SUMMARIZE_URL` - Endpoint for summarization model
   - `SUMMARIZE_MODEL` - Model identifier

## Data Source Registration

The embedding model is automatically registered as a Hugr data source during the load process using environment variable substitution:

```graphql
mutation {
  core {
    insert_data_sources(data: {
      name: "emb_gemma"
      type: "embedding"
      description: "Text embedding model (EmbeddingGemma 308M)"
      path: "[$EMBEDDINGS_URL]?model=[$EMBEDDINGS_MODEL]"
    }) {
      name
    }
  }
}
```

**Note**: The `[$ENV_NAME]` syntax is replaced with actual environment variable values at data source loading time by Hugr.

## Requirements

- Docker Desktop 4.36+ or Docker Engine with AI extension
- GPU support recommended (NVIDIA, AMD, or Apple Silicon)
- Sufficient memory for model inference (configured in docker-compose.yml)

## Usage

### Starting Services

Models are automatically pulled and started when you run:

```bash
./lde/scripts/start.sh
```

### Accessing Models

Models are accessed through Hugr's GraphQL API:

```graphql
# Example: Get embeddings for text
query {
  function {
    emb_gemma {
      embed(input: "your text here") {
        embedding
      }
    }
  }
}
```

### Checking Model Status

```bash
# Check if models are running
docker compose -f lde/docker-compose.yml ps

# View model environment variables
docker exec lde-hugr printenv | grep -E "(EMBEDDINGS|SUMMARIZE)"
```

## Fallback Configuration

If Docker models are not available, the system falls back to using `host.docker.internal:1234`, expecting a local model server (like LM Studio or Ollama) running on the host.

## Troubleshooting

### Models Not Loading
```bash
# Check Docker AI extension
docker info | grep -i models

# Pull models manually
docker model pull ai/embeddinggemma:latest
docker model pull ai/gpt-oss:latest
```

### Memory Issues
Adjust memory limits in `docker-compose.yml` under the `hugr` service's `deploy.resources.limits`.

### Performance
- Models use GPU acceleration when available
- First inference may be slow (model loading)
- Subsequent requests benefit from model caching

## References

- [Docker Models Documentation](https://docs.docker.com/ai/compose/models-and-compose/)
- [Docker AI Extension](https://docs.docker.com/desktop/extensions/marketplace/)
