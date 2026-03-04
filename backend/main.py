from fastapi import FastAPI, Depends, HTTPException, Request
from fastapi.staticfiles import StaticFiles
from fastapi.responses import HTMLResponse, RedirectResponse
from fastapi.middleware.cors import CORSMiddleware
from sqlalchemy.orm import Session
import os
from contextlib import asynccontextmanager

from dotenv import load_dotenv
load_dotenv()

import models, database, api_routes, admin_routes
from database import engine, get_db

# Create the database tables
models.Base.metadata.create_all(bind=engine)

# To simulate the insertion of initial plans
def seed_database(db: Session):
    # Check if plans exist
    plan_count = db.query(models.Plano).count()
    if plan_count == 0:
        print("🌱 Seeding initial plans...")
        plan1 = models.Plano(nome="Acesso 24 Horas", preco=15.00, duracao_minutos=1440, ativo=True)
        # Assuming Horas Personalizadas will be generated dynamically, but we add a base record
        plan2 = models.Plano(nome="Horas Personalizadas (Base 1h)", preco=5.00, duracao_minutos=60, ativo=True)
        db.add(plan1)
        db.add(plan2)
        db.commit()

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup actions
    db = database.SessionLocal()
    seed_database(db)
    db.close()
    yield
    # Shutdown actions

app = FastAPI(title="Astrolink Backend API", lifespan=lifespan)

# CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Determine the absolute path to the frontend folder (one level up from backend)
BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
FRONTEND_DIR = os.path.join(BASE_DIR, "frontend")

# Serve UI endpoints
@app.get("/login", response_class=HTMLResponse)
async def serve_portal():
    with open(os.path.join(FRONTEND_DIR, "index.html"), "r", encoding="utf-8") as f:
        return f.read()

@app.get("/admin", include_in_schema=False)
async def redirect_admin():
    return RedirectResponse(url="/admin/")

# Include the business logic routes
app.include_router(api_routes.router)
app.include_router(admin_routes.router)

# Mount frontend files (admin and captive portal)
app.mount("/admin", StaticFiles(directory=os.path.join(FRONTEND_DIR, "admin"), html=True), name="admin_static")
app.mount("/", StaticFiles(directory=FRONTEND_DIR, html=True), name="static")

if __name__ == "__main__":
    import uvicorn
    uvicorn.run("backend.main:app", host="0.0.0.0", port=8000, reload=True)
