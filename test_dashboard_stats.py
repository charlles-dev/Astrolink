import sys
import os

# Adiciona o diretório backend ao PYTHONPATH para importar models
sys.path.append(os.path.abspath('backend'))

from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from datetime import datetime
from sqlalchemy import func
import models

engine = create_engine('sqlite:///db/hotspot.db')
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
db = SessionLocal()

hoje = datetime.now().date()

try:
    print("Testing faturamento...")
    faturamento_pix = db.query(func.sum(models.Plano.preco))\
        .join(models.TransacaoPix, models.Plano.id == models.TransacaoPix.plano_id)\
        .filter(models.TransacaoPix.status == 'pago')\
        .filter(func.date(models.TransacaoPix.data_criacao) == hoje)\
        .scalar() or 0.0
    print(f"faturamento: {faturamento_pix}")
except Exception as e:
    import traceback
    traceback.print_exc()

try:
    print("Testing online users...")
    now = datetime.now()
    usuarios_online = db.query(models.Usuario).filter(models.Usuario.status == 'ativo', models.Usuario.fim_acesso > now).count()
    print(f"online: {usuarios_online}")
except Exception as e:
    import traceback
    traceback.print_exc()
