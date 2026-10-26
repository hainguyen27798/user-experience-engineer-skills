from fastapi import FastAPI

app = FastAPI(title='{{API_MODULE}}')


@app.get('/api/v1/health')
async def health() -> dict[str, str]:
    return {'status': 'ok'}
