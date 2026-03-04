import os
import sys
import uvicorn

def main():
    # Ensure we are in the root directory
    base_dir = os.path.dirname(os.path.abspath(__file__))
    os.chdir(base_dir)
    
    # Add the backend directory to Python path
    sys.path.append(os.path.join(base_dir, "backend"))
    
    print("🚀 Iniciando o servidor Astrolink...")
    print("🌐 Portal de Clientes: http://127.0.0.1:8000")
    print("⚙️  Painel Admin: http://127.0.0.1:8000/admin")
    
    # Run the FastAPI app using uvicorn
    uvicorn.run("backend.main:app", host="0.0.0.0", port=8000, reload=True)

if __name__ == "__main__":
    main()
