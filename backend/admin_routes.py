import os
import shutil
import time
from datetime import datetime, timedelta
from typing import Dict, Any, List, Optional

from fastapi import APIRouter, Depends, HTTPException, Body, Request, UploadFile, File, Form
from fastapi.responses import Response, FileResponse
from fastapi.security import OAuth2PasswordBearer
import jwt
from sqlalchemy.orm import Session
from sqlalchemy import func

import models, database

# Load environment variables manually to avoid dependency issues if dotenv fails 
# (though we recommend using it, let's keep it safe)
try:
    from dotenv import load_dotenv
    load_dotenv()
except ImportError:
    pass

router = APIRouter(prefix="/api/admin")

ADMIN_USER = os.getenv("ADMIN_USER", "astrolink")
ADMIN_PASS = os.getenv("ADMIN_PASS", "123456")
SECRET_KEY = os.getenv("JWT_SECRET", "super-secret-key-change-in-production")
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 60 * 24  # 1 day

oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/api/admin/login")

def create_access_token(data: dict, expires_delta: timedelta | None = None):
    to_encode = data.copy()
    if expires_delta:
        expire = datetime.utcnow() + expires_delta
    else:
        expire = datetime.utcnow() + timedelta(minutes=15)
    to_encode.update({"exp": expire})
    encoded_jwt = jwt.encode(to_encode, SECRET_KEY, algorithm=ALGORITHM)
    return encoded_jwt

# Dependency for authenticating admin routes
async def get_current_admin(request: Request):
    # Try header first (for API calls)
    token = request.headers.get("Authorization")
    if token and token.startswith("Bearer "):
        token = token.split(" ")[1]
    else:
        # Fallback to cookies (ideal for SSR / simple UI)
        token = request.cookies.get("admin_token")
    
    if not token:
        raise HTTPException(status_code=401, detail="Não autenticado")
    
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        username: str = payload.get("sub")
        if username != ADMIN_USER:
            raise HTTPException(status_code=401, detail="Credenciais inválidas")
    except jwt.PyJWTError:
        raise HTTPException(status_code=401, detail="Token inválido")
    
    return username


@router.post("/login")
def login(payload: Dict[str, Any] = Body(...)):
    username = payload.get("username")
    password = payload.get("password")

    if username == ADMIN_USER and password == ADMIN_PASS:
        access_token_expires = timedelta(minutes=ACCESS_TOKEN_EXPIRE_MINUTES)
        access_token = create_access_token(
            data={"sub": username}, expires_delta=access_token_expires
        )
        return {"access_token": access_token, "token_type": "bearer"}
    else:
        raise HTTPException(status_code=401, detail="Usuário ou senha incorretos")

@router.get("/dashboard-stats")
def dashboard_stats(admin: str = Depends(get_current_admin), db: Session = Depends(database.get_db)):
    hoje = datetime.now().date()
    
    # 1. Total faturamento PIX hoje
    faturamento_pix = db.query(func.sum(models.Plano.preco))\
        .join(models.TransacaoPix, models.Plano.id == models.TransacaoPix.plano_id)\
        .filter(models.TransacaoPix.status == 'pago')\
        .filter(func.date(models.TransacaoPix.data_criacao) == hoje)\
        .scalar() or 0.0

    # 2. Usuários Online Agora
    now = datetime.now()
    usuarios_online = db.query(models.Usuario).filter(models.Usuario.status == 'ativo', models.Usuario.fim_acesso > now).count()

    # 3. Vouchers Disponíveis
    vouchers_disponiveis = db.query(models.Voucher).filter(models.Voucher.status == 'disponivel').count()

    return {
        "faturamento_hoje": faturamento_pix,
        "online_agora": usuarios_online,
        "vouchers_disponiveis": vouchers_disponiveis
    }

@router.post("/gerar-vouchers")
def gerar_vouchers(payload: Dict[str, Any] = Body(...), admin: str = Depends(get_current_admin), db: Session = Depends(database.get_db)):
    duracao_minutos = payload.get("duracao_minutos", 1440)  # Default 24h
    quantidade = payload.get("quantidade", 1)
    is_universal = payload.get("is_universal", False)
    max_uses = payload.get("max_uses", 1)
    nome = payload.get("nome", None)

    valor_pago = payload.get("valor_pago", 0.0)

    # Sanitize input
    if quantidade > 100:
        quantidade = 100
        
    import random
    import string
    
    # Avoid ambiguous characters (0, O, I, L, 1)
    chars = "ABCDEFGHJKMNPQRTUVWXYZ2346789"

    gerados = []
    
    if is_universal:
        quantidade = 1 # Universal is only 1 code
        if not nome or nome.strip() == "":
            part1 = ''.join(random.choices(chars, k=3))
            part2 = ''.join(random.choices(chars, k=3))
            codigo = f"UNIV-{part1}-{part2}"
        else:
            codigo = nome.strip().upper().replace(" ", "-")
        
        # Check if already exists
        exists = db.query(models.Voucher).filter(models.Voucher.codigo == codigo).first()
        if exists:
            # Maybe append random chars if exists
            part1 = ''.join(random.choices(chars, k=2))
            codigo = f"{codigo}-{part1}"
            
        novo_voucher = models.Voucher(
            codigo=codigo,
            nome=nome,
            duracao_minutos=duracao_minutos,
            valor_pago=valor_pago,
            status="disponivel",
            is_universal=True,
            max_uses=max_uses,
            uses_count=0
        )
        db.add(novo_voucher)
        gerados.append(codigo)
    else:    
        for _ in range(quantidade):
            part1 = ''.join(random.choices(chars, k=3))
            part2 = ''.join(random.choices(chars, k=3))
            codigo = f"{part1}-{part2}"
            
            novo_voucher = models.Voucher(
                codigo=codigo,
                duracao_minutos=duracao_minutos,
                valor_pago=valor_pago,
                status="disponivel",
                is_universal=False,
                max_uses=1,
                uses_count=0
            )
            db.add(novo_voucher)
            gerados.append(codigo)
        
    db.commit()

    return {"mensagem": f"{quantidade} vouchers gerados com sucesso!", "vouchers": gerados}

@router.get("/vouchers")
def list_vouchers(admin: str = Depends(get_current_admin), db: Session = Depends(database.get_db)):
    # Order by newest first, limit to 50 for performance
    vouchers = db.query(models.Voucher).order_by(models.Voucher.data_criacao.desc()).limit(50).all()
    
    result = []
    for v in vouchers:
        duracao_format = f"{v.duracao_minutos // 60}h" if v.duracao_minutos >= 60 else f"{v.duracao_minutos}m"
        result.append({
            "id": v.id,
            "codigo": v.codigo,
            "duracao": duracao_format,
            "status": v.status,
            "is_universal": getattr(v, "is_universal", False),
            "max_uses": getattr(v, "max_uses", 1),
            "uses_count": getattr(v, "uses_count", 0),
            "criado_em": v.data_criacao.strftime("%d/%m/%Y %H:%M") if v.data_criacao else "-",
            "usado_em": v.data_uso.strftime("%d/%m/%Y %H:%M") if v.data_uso else "-",
            "usado_por": v.mac_address_usado or "-"
        })
    return result

@router.get("/usuarios-ativos")
def usuarios_ativos(admin: str = Depends(get_current_admin), db: Session = Depends(database.get_db)):
    now = datetime.now()
    usuarios = db.query(models.Usuario).filter(models.Usuario.status == 'ativo', models.Usuario.fim_acesso > now).all()
    
    result = []
    for u in usuarios:
        # Calculate remaining time
        if u.fim_acesso:
            delta = u.fim_acesso - now
            hours, remainder = divmod(delta.seconds, 3600)
            minutes, _ = divmod(remainder, 60)
            tempo_restante = f"{delta.days}d {hours}h {minutes}m" if delta.days > 0 else f"{hours}h {minutes}m"
        else:
            tempo_restante = "Ilimitado"

        result.append({
            "id": u.id,
            "mac_address": u.mac_address,
            "ip_atual": u.ip_atual or "Desconhecido",
            "tempo_restante": tempo_restante,
            "fim_acesso": u.fim_acesso.strftime("%d/%m/%Y %H:%M") if u.fim_acesso else "-"
        })
    return result

@router.post("/derrubar-usuario")
def derrubar_usuario(payload: Dict[str, Any] = Body(...), admin: str = Depends(get_current_admin), db: Session = Depends(database.get_db)):
    mac_address = payload.get("mac_address")
    
    if not mac_address:
        raise HTTPException(status_code=400, detail="MAC Address não fornecido")
        
    usuario = db.query(models.Usuario).filter(models.Usuario.mac_address == mac_address).first()
    if not usuario:
        raise HTTPException(status_code=404, detail="Usuário não encontrado")
        
    # Revoke access in DB
    usuario.status = "bloqueado"
    usuario.fim_acesso = datetime.now()
    db.commit()
    
    # FUTURE: SIMULATE OPENNDS KICK
    print(f"🔪 [OpenNDS SIMULATOR] ndsctl deauth {mac_address} executed by ADMIN.")
    
    # Log the kick
    db.add(models.Log(mac_address=mac_address, event_type="USER_KICKED", description=f"Admin derrubou acesso de {mac_address}"))
    db.commit()
    
    return {"mensagem": f"Usuário {mac_address} derrubado com sucesso"}

# ================= GERENCIADOR DE PLANOS =================

@router.get("/planos")
def admin_get_planos(admin: str = Depends(get_current_admin), db: Session = Depends(database.get_db)):
    try:
        planos = db.query(models.Plano).order_by(models.Plano.id.asc()).all()
        return [{"id": p.id, "nome": p.nome, "duracao_minutos": p.duracao_minutos, "max_devices": p.max_devices, "preco": p.preco, "ativo": p.ativo} for p in planos]
    except Exception as e:
        import traceback
        with open("error_log.txt", "w") as f:
            f.write(traceback.format_exc())
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/planos")
def admin_create_plano(payload: Dict[str, Any] = Body(...), admin: str = Depends(get_current_admin), db: Session = Depends(database.get_db)):
    novo_plano = models.Plano(
        nome=payload.get("nome"),
        duracao_minutos=payload.get("duracao_minutos"),
        max_devices=payload.get("max_devices", 1),
        preco=payload.get("preco"),
        ativo=payload.get("ativo", True)
    )
    db.add(novo_plano)
    db.commit()
    return {"mensagem": "Plano criado com sucesso!"}

@router.put("/planos/{plano_id}")
def admin_update_plano(plano_id: int, payload: Dict[str, Any] = Body(...), admin: str = Depends(get_current_admin), db: Session = Depends(database.get_db)):
    plano = db.query(models.Plano).filter(models.Plano.id == plano_id).first()
    if not plano:
        raise HTTPException(status_code=404, detail="Plano não encontrado")
    
    if "nome" in payload:
        plano.nome = payload["nome"]
    if "duracao_minutos" in payload:
        plano.duracao_minutos = payload["duracao_minutos"]
    if "max_devices" in payload:
        plano.max_devices = payload["max_devices"]
    if "preco" in payload:
        plano.preco = payload["preco"]
    if "ativo" in payload:
        plano.ativo = payload["ativo"]
        
    db.commit()
    return {"mensagem": "Plano atualizado com sucesso!"}

@router.patch("/planos/{plano_id}/toggle")
def admin_toggle_plano(plano_id: int, admin: str = Depends(get_current_admin), db: Session = Depends(database.get_db)):
    plano = db.query(models.Plano).filter(models.Plano.id == plano_id).first()
    if not plano:
        raise HTTPException(status_code=404, detail="Plano não encontrado")
    
    plano.ativo = not plano.ativo
    db.commit()
    return {"mensagem": f"Plano {'ativado' if plano.ativo else 'desativado'} com sucesso!"}

@router.delete("/planos/{plano_id}")
def admin_delete_plano(plano_id: int, admin: str = Depends(get_current_admin), db: Session = Depends(database.get_db)):
    plano = db.query(models.Plano).filter(models.Plano.id == plano_id).first()
    if not plano:
        raise HTTPException(status_code=404, detail="Plano não encontrado")
        
    try:
        db.delete(plano)
        db.commit()
        return {"mensagem": "Plano excluído com sucesso!"}
    except Exception as e:
        db.rollback()
        raise HTTPException(status_code=400, detail="Não foi possível excluir o plano, provavelmente há transações vinculadas a ele.")

# ================= ADVANCED FEATURES =================

@router.get("/relatorios/financeiro")
def relatorio_financeiro(admin: str = Depends(get_current_admin), db: Session = Depends(database.get_db)):
    # Calculate revenue for the current month
    now = datetime.now()
    primeiro_dia_mes = now.replace(day=1, hour=0, minute=0, second=0, microsecond=0)
    
    # PIX Revenue (paid)
    receita_pix = db.query(func.sum(models.Plano.preco))\
        .join(models.TransacaoPix, models.Plano.id == models.TransacaoPix.plano_id)\
        .filter(models.TransacaoPix.status == 'pago')\
        .filter(models.TransacaoPix.data_criacao >= primeiro_dia_mes)\
        .scalar() or 0.0
        
    # Voucher Revenue (presumptive cash, available or used)
    receita_voucher = db.query(func.sum(models.Voucher.valor_pago))\
        .filter(models.Voucher.data_criacao >= primeiro_dia_mes)\
        .scalar() or 0.0
        
    return {
        "mes_atual": now.strftime("%Y-%m"),
        "receita_pix": round(receita_pix, 2),
        "receita_voucher": round(receita_voucher, 2),
        "total": round(receita_pix + receita_voucher, 2)
    }

@router.get("/relatorios/exportar-csv")
def exportar_csv(admin: str = Depends(get_current_admin), db: Session = Depends(database.get_db)):
    import csv
    import io
    
    output = io.StringIO()
    writer = csv.writer(output, delimiter=';')
    
    # Headers
    writer.writerow(["Data", "Tipo", "Valor (R$)", "Status", "MAC Address / Codigo"])
    
    # PIX Transactions
    transacoes_pix = db.query(models.TransacaoPix, models.Plano).join(models.Plano).all()
    for t, p in transacoes_pix:
        date_str = t.data_criacao.strftime("%Y-%m-%d %H:%M:%S") if t.data_criacao else ""
        writer.writerow([date_str, "PIX (Online)", f"{p.preco:.2f}", t.status, t.mac_address])
        
    # Vouchers
    vouchers = db.query(models.Voucher).all()
    for v in vouchers:
        date_str = v.data_criacao.strftime("%Y-%m-%d %H:%M:%S") if v.data_criacao else ""
        writer.writerow([date_str, "VOUCHER (Dinheiro)", f"{v.valor_pago:.2f}", v.status, v.codigo])
        
    response = Response(content=output.getvalue(), media_type="text/csv")
    response.headers["Content-Disposition"] = f"attachment; filename=relatorio_financeiro_{datetime.now().strftime('%Y%m%d')}.csv"
    return response

@router.get("/walled-garden")
def list_walled_garden(admin: str = Depends(get_current_admin), db: Session = Depends(database.get_db)):
    rules = db.query(models.WalledGardenRule).all()
    return rules

@router.post("/walled-garden")
def add_walled_garden(payload: dict = Body(...), admin: str = Depends(get_current_admin), db: Session = Depends(database.get_db)):
    tipo = payload.get("tipo", "IP") # IP or URL
    valor = payload.get("valor")
    descricao = payload.get("descricao", "")
    
    if not valor:
        raise HTTPException(status_code=400, detail="Valor é obrigatório")
        
    rule = models.WalledGardenRule(tipo=tipo, valor=valor, descricao=descricao)
    db.add(rule)
    db.commit()
    db.refresh(rule)
    
    return {"mensagem": "Regra adicionada com sucesso", "regra": rule}

@router.delete("/walled-garden/{rule_id}")
def delete_walled_garden(rule_id: int, admin: str = Depends(get_current_admin), db: Session = Depends(database.get_db)):
    rule = db.query(models.WalledGardenRule).filter(models.WalledGardenRule.id == rule_id).first()
    if not rule:
        raise HTTPException(status_code=404, detail="Regra não encontrada")
        
    db.delete(rule)
    db.commit()
    return {"mensagem": "Regra removida"}

@router.get("/logs")
def list_logs(
    limit: int = 50, 
    start_date: Optional[str] = None, 
    end_date: Optional[str] = None, 
    event_type: Optional[str] = None, 
    admin: str = Depends(get_current_admin), 
    db: Session = Depends(database.get_db)
):
    query = db.query(models.Log)

    if start_date:
        query = query.filter(models.Log.created_at >= start_date)
    if end_date:
        query = query.filter(models.Log.created_at <= f"{end_date} 23:59:59")
    if event_type and event_type != "ALL":
        query = query.filter(models.Log.event_type.like(f"%{event_type}%"))

    # Most recent first
    logs = query.order_by(models.Log.created_at.desc()).limit(limit).all()
    return logs

# ================= NOTIFICATIONS (Polling) =================

@router.get("/notifications")
def get_notifications(admin: str = Depends(get_current_admin), db: Session = Depends(database.get_db)):
    # A simple mock returning recent important logs as notifications
    # In a real scenario, this could query a specific notifications table
    # We will fetch 'FAIL' or specific system events from the last 24h
    now = datetime.now()
    yesterday = now - timedelta(days=1)
    
    important_logs = db.query(models.Log)\
        .filter(models.Log.created_at >= yesterday)\
        .filter(models.Log.event_type.in_(['AUTH_FAIL', 'SYSTEM_ERROR', 'ROUTER_DISCONNECT']))\
        .order_by(models.Log.created_at.desc())\
        .limit(5).all()
        
    return important_logs

# ================= SYSTEM SETTINGS =================

@router.get("/system-settings")
def get_settings(admin: str = Depends(get_current_admin), db: Session = Depends(database.get_db)):
    settings = db.query(models.SystemSettings).all()
    # Format as { key: value }
    return {s.key: s.value for s in settings}

@router.post("/system-settings")
async def update_settings(
    request: Request,
    admin: str = Depends(get_current_admin),
    db: Session = Depends(database.get_db)
):
    form_data = await request.form()
    
    # Extract file if it exists
    file: UploadFile = form_data.get("file")
    
    # Extract all other key-value pairs to assemble the payload
    payload = {}
    for key, value in form_data.items():
        if key != "file" and isinstance(value, str):
            payload[key] = value

    # Handle optional logo upload right here
    if file and file.filename:
        if not file.content_type.startswith("image/"):
            raise HTTPException(status_code=400, detail="Formato de arquivo inválido. Deve ser uma imagem.")

        ext = os.path.splitext(file.filename)[1]
        filename = f"logo_{int(time.time())}{ext}"
        
        base_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
        upload_dir = os.path.join(base_dir, "frontend", "public", "uploads")
        os.makedirs(upload_dir, exist_ok=True)
        
        file_path = os.path.join(upload_dir, filename)
        with open(file_path, "wb") as buffer:
            shutil.copyfileobj(file.file, buffer)
            
        # Overwrite the payload URL with the newly uploaded local asset path
        payload["LOGO_URL"] = f"/public/uploads/{filename}"

    # Proceed to save all settings to DB
    for key, value in payload.items():
        if value is None:
            continue
        setting = db.query(models.SystemSettings).filter(models.SystemSettings.key == key).first()
        if setting:
            setting.value = str(value)
        else:
            new_setting = models.SystemSettings(key=key, value=str(value))
            db.add(new_setting)
    
    db.commit()
    # Log the action
    new_log = models.Log(event_type="SETTINGS_UPDATE", description=f"Admin updated system settings.")
    db.add(new_log)
    db.commit()
    
    # Return updated url if uploaded so front-end can update without refresh
    return {
        "status": "success", 
        "mensagem": "Configurações Globais salvas com sucesso.", 
        "logo_url": payload.get("LOGO_URL", form_data.get("LOGO_URL"))
    }

# ================= USER DETAILS =================
@router.get("/usuario/{mac}")
def get_user_details(mac: str, admin: str = Depends(get_current_admin), db: Session = Depends(database.get_db)):
    import urllib.parse
    mac = urllib.parse.unquote(mac).upper()
    
    now = datetime.now()
    usuario = db.query(models.Usuario).filter(models.Usuario.mac_address == mac, models.Usuario.status == 'ativo').first()
    
    sessao_atual = None
    if usuario and usuario.fim_acesso and usuario.fim_acesso > now:
        delta = usuario.fim_acesso - now
        horas, restos = divmod(delta.total_seconds(), 3600)
        minutos, _ = divmod(restos, 60)
        sessao_atual = {
            "ip_atual": usuario.ip_atual,
            "tempo_restante": f"{int(horas)}h {int(minutos)}m",
            "fim_acesso": usuario.fim_acesso.strftime("%d/%m/%Y %H:%M")
        }
    
    logs = db.query(models.Log)\
             .filter(models.Log.mac_address == mac)\
             .order_by(models.Log.created_at.desc())\
             .limit(10).all()
    
    historico_eventos = [{
        "data_hora": log.created_at.strftime("%d/%m/%Y %H:%M"),
        "evento": log.event_type,
        "descricao": log.description
    } for log in logs]
    
    return {
        "mac_address": mac,
        "sessao_atual": sessao_atual,
        "historico_eventos": historico_eventos,
        "historico_vouchers": [] 
    }

# ================= BACKUP & RESTORE =================
import shutil
from pathlib import Path

DB_FILE_PATH = "./db/hotspot.db"

@router.get("/system/backup")
def backup_database(admin: str = Depends(get_current_admin)):
    if not os.path.exists(DB_FILE_PATH):
        raise HTTPException(status_code=404, detail="Banco de dados não encontrado")
    
    return FileResponse(
        path=DB_FILE_PATH,
        filename=f"astrolink_backup_{datetime.now().strftime('%Y%m%d_%H%M%S')}.db",
        media_type="application/octet-stream"
    )

@router.post("/system/restore")
def restore_database(file: UploadFile = File(...), admin: str = Depends(get_current_admin)):
    if not file.filename.endswith('.db'):
        raise HTTPException(status_code=400, detail="O arquivo deve ser um backup .db válido")
    
    try:
        temp_path = f"./db/temp_restore_{datetime.now().strftime('%s')}.db"
        with open(temp_path, "wb") as buffer:
            shutil.copyfileobj(file.file, buffer)
        
        if os.path.exists(DB_FILE_PATH):
            import time
            backup_old = f"{DB_FILE_PATH}.old.{int(time.time())}"
            shutil.copy2(DB_FILE_PATH, backup_old)
            
        shutil.move(temp_path, DB_FILE_PATH)
        return {"mensagem": "Banco de dados restaurado com sucesso! Recomendamos reiniciar o serviço (backend)."}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

# ================= DIAGNÓSTICOS DE REDE =================
import subprocess
import platform

@router.post("/system/ping")
def ping_target(payload: dict = Body(...), admin: str = Depends(get_current_admin)):
    target = payload.get("target")
    if not target:
        raise HTTPException(status_code=400, detail="O destino (IP ou Host) é obrigatório")
        
    try:
        param = '-n' if platform.system().lower() == 'windows' else '-c'
        command = ['ping', param, '4', target]
        
        output = subprocess.check_output(command, stderr=subprocess.STDOUT, text=True, timeout=10)
        return {"resultado": output}
    except subprocess.TimeoutExpired:
        return {"resultado": "Timeout. O host não respondeu a tempo."}
    except subprocess.CalledProcessError as e:
        return {"resultado": f"Falha no Ping:\\n{e.output}"}
    except Exception as e:
        return {"resultado": f"Erro interno: {str(e)}"}

# ================= BLACKLIST / BLOQUEIO DE MAC =================
from pydantic import BaseModel

class GenericResponse(BaseModel):
    mensagem: str

class BlacklistCreate(BaseModel):
    mac_address: str
    motivo: str = None

class ContaPersistenteCreate(BaseModel):
    mac_address: str
    nome: str
    telefone: str = None
    email: str = None
    observacoes: str = None
    data_renovacao: str = None # Format YYYY-MM-DD ou deixado vazio
    ativo: bool = True

@router.get("/blacklist")
def get_blacklist(db: Session = Depends(database.get_db), admin: str = Depends(get_current_admin)):
    records = db.query(models.BlacklistMAC).order_by(models.BlacklistMAC.data_bloqueio.desc()).all()
    return [{
        "id": r.id,
        "mac_address": r.mac_address,
        "motivo": r.motivo,
        "data_bloqueio": r.data_bloqueio.strftime("%d/%m/%Y %H:%M") if r.data_bloqueio else ""
    } for r in records]

@router.post("/blacklist")
def add_to_blacklist(payload: BlacklistCreate, db: Session = Depends(database.get_db), admin: str = Depends(get_current_admin)):
    mac = payload.mac_address.upper()
    
    # Check if already blacklisted
    existing = db.query(models.BlacklistMAC).filter(models.BlacklistMAC.mac_address == mac).first()
    if existing:
        raise HTTPException(status_code=400, detail="Este MAC já está na blacklist.")
        
    new_ban = models.BlacklistMAC(mac_address=mac, motivo=payload.motivo)
    db.add(new_ban)
    
    # Also update user status if it exists
    d_user = db.query(models.Usuario).filter(models.Usuario.mac_address == mac).first()
    if d_user:
        d_user.status = "banido"
        
    db.commit()
    
    # Try to disconnect from router
    try:
        from core import derrubar_usuario_mikrotik
        derrubar_usuario_mikrotik(mac)
    except Exception as e:
        pass # If router fails, it's fine, at least db is updated

    return {"mensagem": f"MAC {mac} adicionado à blacklist permanente."}

@router.delete("/blacklist/{mac_address}")
def remove_from_blacklist(mac_address: str, db: Session = Depends(database.get_db), admin: str = Depends(get_current_admin)):
    mac = mac_address.upper()
    banned = db.query(models.BlacklistMAC).filter(models.BlacklistMAC.mac_address == mac).first()
    if not banned:
        raise HTTPException(status_code=404, detail="MAC não encontrado na blacklist.")
        
    db.delete(banned)
    
    # Keep user blocked but not banned, so they can buy again
    d_user = db.query(models.Usuario).filter(models.Usuario.mac_address == mac).first()
    if d_user:
        d_user.status = "bloqueado"

    db.commit()
    return {"mensagem": f"MAC {mac} removido da blacklist."}

# ==========================================
# GESTÃO DE CONTAS PERSISTENTES (MENSALISTAS)
# ==========================================

@router.get("/contas-fixas")
def list_contas_fixas(db: Session = Depends(database.get_db), admin: str = Depends(get_current_admin)):
    contas = db.query(models.ContaPersistente).all()
    res = []
    for c in contas:
        res.append({
            "id": c.id,
            "mac_address": c.mac_address,
            "nome": c.nome,
            "telefone": c.telefone,
            "email": c.email,
            "observacoes": c.observacoes,
            "data_renovacao": c.data_renovacao.strftime("%Y-%m-%d") if c.data_renovacao else None,
            "ativo": c.ativo,
            "criado_em": c.criado_em.strftime("%Y-%m-%d %H:%M:%S") if c.criado_em else None
        })
    return res

@router.post("/contas-fixas")
def save_conta_fixa(payload: ContaPersistenteCreate, db: Session = Depends(database.get_db), admin: str = Depends(get_current_admin)):
    mac = payload.mac_address.upper()
    
    # Check if user exists in the generic table, if not create one
    usuario = db.query(models.Usuario).filter(models.Usuario.mac_address == mac).first()
    if not usuario:
        usuario = models.Usuario(mac_address=mac, status="ativo")
        db.add(usuario)
        db.commit()
        db.refresh(usuario)
        
    # Check if fixed account exists
    conta = db.query(models.ContaPersistente).filter(models.ContaPersistente.mac_address == mac).first()
    
    renovacao_dt = None
    if payload.data_renovacao:
        try:
            renovacao_dt = datetime.strptime(payload.data_renovacao, "%Y-%m-%d")
        except:
            pass
            
    if conta:
        # Update
        conta.nome = payload.nome
        conta.telefone = payload.telefone
        conta.email = payload.email
        conta.observacoes = payload.observacoes
        conta.data_renovacao = renovacao_dt
        conta.ativo = payload.ativo
    else:
        # Create
        conta = models.ContaPersistente(
            mac_address=mac,
            nome=payload.nome,
            telefone=payload.telefone,
            email=payload.email,
            observacoes=payload.observacoes,
            data_renovacao=renovacao_dt,
            ativo=payload.ativo
        )
        db.add(conta)
        
    # If active and it's a fixed account, we can set fim_acesso far into the future (or link to data_renovacao)
    if payload.ativo:
        usuario.status = "ativo"
        # If user pays monthly, set their expiration to renovacao date + 1 day
        if renovacao_dt:
            usuario.fim_acesso = renovacao_dt + timedelta(days=1)
        else:
            # Persistent forever
            usuario.fim_acesso = datetime.utcnow() + timedelta(days=365*10)
    else:
        # Block them immediately if inactive
        usuario.status = "bloqueado"
        usuario.fim_acesso = datetime.utcnow() - timedelta(days=1)
        # Try to disconnect from router
        try:
            from core import derrubar_usuario_mikrotik
            derrubar_usuario_mikrotik(mac)
        except:
            pass

    db.commit()
    return {"mensagem": f"Conta Fixa para {mac} salva com sucesso!"}

@router.delete("/contas-fixas/{mac_address}")
def delete_conta_fixa(mac_address: str, db: Session = Depends(database.get_db), admin: str = Depends(get_current_admin)):
    mac = mac_address.upper()
    conta = db.query(models.ContaPersistente).filter(models.ContaPersistente.mac_address == mac).first()
    if not conta:
        raise HTTPException(status_code=404, detail="Conta fixa não encontrada.")
    
    db.delete(conta)
    
    # Revoke access as normal user
    usuario = db.query(models.Usuario).filter(models.Usuario.mac_address == mac).first()
    if usuario:
        usuario.status = "bloqueado"
        usuario.fim_acesso = datetime.utcnow() - timedelta(days=1)
        
    db.commit()
    
    # Try to disconnect
    try:
        from core import derrubar_usuario_mikrotik
        derrubar_usuario_mikrotik(mac)
    except:
        pass
        
    return {"mensagem": f"Conta fixa ({mac}) removida e acesso cancelado."}

class QoSRequest(BaseModel):
    mac_address: str
    limit: str   # Ex: "10M", "50M" or "" to reset

@router.post("/apply-qos")
def apply_qos(req: QoSRequest, db: Session = Depends(database.get_db), admin: str = Depends(get_current_admin)):
    """
    Aplica limites de banda (QoS) em um usuário via OpenWrt.
    Atualmente enviando log e simulando comunicação SSH com o roteador.
    """
    mac = req.mac_address.upper()
    limit = req.limit.strip()

    # Buscamos configurações do Roteador (ex Mikrotik, agora OpenWrt)
    router_ip = db.query(models.SystemSettings).filter(models.SystemSettings.key == "ROUTER_IP").first()
    
    if not limit:
        action_msg = f"Limite de banda removido para {mac}."
        print(f"📡 [SSH -> OpenWrt] Removendo regras QoS para o MAC {mac}")
    else:
        action_msg = f"Banda limitada em {limit} para {mac}."
        print(f"📡 [SSH -> OpenWrt] Conectando ao roteador {router_ip.value if router_ip else '192.168.1.1'}...")
        print(f"📡 [SSH -> OpenWrt] Executando: tc qdisc add dev br-lan... rate {limit} para MAC {mac}")

    # Registra no log de auditoria
    db.add(models.Log(
        mac_address=mac,
        event_type="SYSTEM",
        description=action_msg
    ))
    db.commit()

    return {"mensagem": action_msg}
