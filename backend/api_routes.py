from fastapi import APIRouter, Depends, HTTPException, Body
from sqlalchemy.orm import Session
from datetime import datetime, timedelta
from typing import Dict, Any
import uuid
import os
import mercadopago

import models, database

router = APIRouter(prefix="/api")

# Initialize Mercado Pago SDK
import models, database

router = APIRouter(prefix="/api")


@router.get("/planos")
def get_planos(db: Session = Depends(database.get_db)):
    """ Returns active plans to the frontend """
    planos = db.query(models.Plano).filter(models.Plano.ativo == True).all()
    # Convert to format expected by the frontend
    result = []
    for p in planos:
        result.append({
            "id": p.id,
            "name": p.nome,
            "duration": f"{p.duracao_minutos // 60} Horas" if p.duracao_minutos >= 60 else f"{p.duracao_minutos} Minutos",
            "price": f"R$ {p.preco:.2f}".replace(".", ","),
            "description": "Mock description",
            "popular": p.nome == "Acesso 24 Horas",
            "isCustom": p.duracao_minutos == 60 or "Personalizada" in p.nome
        })
    return result

@router.get("/settings")
def get_public_settings(db: Session = Depends(database.get_db)):
    """ Returns public settings for the Captive Portal """
    keys = [
        "LOGO_URL", "BACKGROUND_URL", "PRIMARY_COLOR", "PROVIDER_NAME",
        "THEME_MODE", "SECONDARY_COLOR", "WELCOME_MSG", "FOOTER_TEXT",
        "TERMS_OF_USE", "REDIRECT_URL", "DEFAULT_LANGUAGE",
        "LOGIN_EMAIL_ENABLED", "LOGIN_PURCHASE_ENABLED", "LOGIN_VOUCHER_ENABLED"
    ]
    settings = db.query(models.SystemSettings).filter(models.SystemSettings.key.in_(keys)).all()
    result = {s.key: s.value for s in settings}
    return result

@router.post("/gerar-pix")
def gerar_pix(payload: Dict[str, Any] = Body(...), db: Session = Depends(database.get_db)):
    mac_address = payload.get("mac")
    plano_id = payload.get("plano_id")
    horas_personalizadas = payload.get("custom_hours") # Only used if the plan is the custom one
    nome = payload.get("nome", "Cliente")
    sobrenome = payload.get("sobrenome", "Astrolink")
    cpf = payload.get("cpf")

    if not mac_address or not plano_id:
        raise HTTPException(status_code=400, detail="Missing MAC or plano_id")

    plano = db.query(models.Plano).filter(models.Plano.id == plano_id).first()
    if not plano:
        raise HTTPException(status_code=404, detail="Plano não encontrado")

    # If it's the custom plan, we need to adjust the price temporarily (or create a new record, but we'll adapt dynamically)
    if "Personalizada" in plano.nome and horas_personalizadas:
        plano.duracao_minutos = int(horas_personalizadas) * 60
        plano.preco = 5.00 * int(horas_personalizadas)  # 5 BRL per hour
    
    # 1. Register or get user
    usuario = db.query(models.Usuario).filter(models.Usuario.mac_address == mac_address).first()
    if not usuario:
        usuario = models.Usuario(mac_address=mac_address, status='walled_garden')
        db.add(usuario)
        db.commit()
        db.refresh(usuario)
    
    # 2. Get MercadoPago token from Settings
    mp_token_setting = db.query(models.SystemSettings).filter(models.SystemSettings.key == "MERCADOPAGO_ACCESS_TOKEN").first()
    if not mp_token_setting or not mp_token_setting.value:
        raise HTTPException(status_code=400, detail="Gateway de Pagamento (MercadoPago) não está configurado.")
    
    sdk = mercadopago.SDK(mp_token_setting.value)

    # 3. Generate Real PIX via Mercado Pago
    payer_info = {
        "email": "tecnologia@astrolink.com",
        "first_name": nome,
        "last_name": sobrenome
    }

    if cpf:
        payer_info["identification"] = {
            "type": "CPF",
            "number": cpf
        }

    payment_data = {
        "transaction_amount": float(plano.preco),
        "description": f"Plano Wi-Fi {plano.nome}",
        "payment_method_id": "pix",
        "payer": payer_info
    }
    
    payment_response = sdk.payment().create(payment_data)
    payment = payment_response.get("response")

    if not payment or "id" not in payment:
        raise HTTPException(status_code=500, detail=f"Erro Mercado Pago: {payment_response}")
        
    mp_id = str(payment["id"])
    pix_copia_cola = payment.get("point_of_interaction", {}).get("transaction_data", {}).get("qr_code", "")
    qr_code_base64 = payment.get("point_of_interaction", {}).get("transaction_data", {}).get("qr_code_base64", "")
    
    # The frontend expects a QR code URL or base64. 
    # Let's provide the base64 as a data URI if available, 
    # otherwise fallback to a generic QR generator URL
    import urllib.parse
    qs = urllib.parse.urlencode({"txt": pix_copia_cola})
    qr_code_url = f"data:image/jpeg;base64,{qr_code_base64}" if qr_code_base64 else f"https://api.qrserver.com/v1/create-qr-code/?size=250x250&data={qs}"

    pix_data = {
        "txid": mp_id,
        "pix_copia_cola": pix_copia_cola,
        "qr_code": qr_code_url
    }
    
    pix_data = {
        "txid": mp_id,
        "pix_copia_cola": pix_copia_cola,
        "qr_code": qr_code_url
    }
    
    # 4. Save transaction to DB
    nova_transacao = models.TransacaoPix(
        txid=mp_id,
        mac_address=mac_address,
        plano_id=plano_id,
        status='pendente'
    )
    db.add(nova_transacao)
    
    # Create log
    novo_log = models.Log(
        mac_address=mac_address,
        event_type="PIX_CREATED",
        description=f"Generated PIX out via MP for {plano.duracao_minutos} mins. TXID: {mp_id}"
    )
    db.add(novo_log)
    
    db.commit()

    # 5. OpenNDS trigger to allow access to banking APIs (simulated in logs)
    print(f"🔒 [OpenNDS SIMULATOR] Walled Garden opened for MAC {mac_address} for 10 minutes.")

    return pix_data

@router.get("/status-pix")
def status_pix(txid: str, db: Session = Depends(database.get_db)):
    transacao = db.query(models.TransacaoPix).filter(models.TransacaoPix.txid == txid).first()
    if not transacao:
        raise HTTPException(status_code=404, detail="Transação não encontrada")
    
    # Se já estiver pago localmente, não bate na API
    if transacao.status == "pago":
        return {"status": "pago"}

    # Checar no Mercado Pago
    mp_token_setting = db.query(models.SystemSettings).filter(models.SystemSettings.key == "MERCADOPAGO_ACCESS_TOKEN").first()
    if not mp_token_setting or not mp_token_setting.value:
        raise HTTPException(status_code=400, detail="Gateway MercadoPago não configurado")
    
    sdk = mercadopago.SDK(mp_token_setting.value)

    try:
        mp_status_resp = sdk.payment().get(txid)
        payment_info = mp_status_resp.get("response")
        
        if payment_info and payment_info.get("status") == "approved":
            # Pagamento Aprovado!
            transacao.status = 'pago'
            
            # Update user access
            usuario = db.query(models.Usuario).filter(models.Usuario.mac_address == transacao.mac_address).first()
            plano = db.query(models.Plano).filter(models.Plano.id == transacao.plano_id).first()
            if usuario and plano:
                usuario.status = 'ativo'
                duracao = transacao.custom_duration_minutes if transacao.custom_duration_minutes else plano.duracao_minutos
                usuario.fim_acesso = datetime.now() + timedelta(minutes=duracao)
                
            db.add(models.Log(
                mac_address=transacao.mac_address,
                event_type="AUTH_SUCCESS",
                description=f"PIX Paid via MP. Internet liberated until {usuario.fim_acesso}"
            ))
            db.commit()
            print(f"🔓 [OpenNDS SIMULATOR] Internet Fully Liberated for MAC {transacao.mac_address} until {usuario.fim_acesso}")
            
        elif payment_info and payment_info.get("status") in ["cancelled", "rejected", "refunded"]:
            transacao.status = payment_info.get("status")
            db.commit()

        return {"status": transacao.status}

    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/webhook-mercadopago")
def webhook_mercadopago(payload: Dict[str, Any] = Body(...), db: Session = Depends(database.get_db)):
    """
    Recebe as notificações IPN do Mercado Pago.
    Libera a internet de forma assíncrona caso a aba do cliente tenha sido fechada.
    """
    action = payload.get("action")
    data = payload.get("data", {})
    payment_id = data.get("id")

    if action in ["payment.created", "payment.updated"] and payment_id:
        mp_token_setting = db.query(models.SystemSettings).filter(models.SystemSettings.key == "MERCADOPAGO_ACCESS_TOKEN").first()
        if not mp_token_setting or not mp_token_setting.value:
            return {"status": "ignored", "reason": "No gateway token"}

        sdk = mercadopago.SDK(mp_token_setting.value)
        
        try:
            mp_status_resp = sdk.payment().get(payment_id)
            payment_info = mp_status_resp.get("response")
            
            if payment_info and payment_info.get("status") == "approved":
                txid = str(payment_id)
                transacao = db.query(models.TransacaoPix).filter(models.TransacaoPix.txid == txid).first()
                
                if transacao and transacao.status != 'pago':
                    transacao.status = 'pago'
                    
                    usuario = db.query(models.Usuario).filter(models.Usuario.mac_address == transacao.mac_address).first()
                    plano = db.query(models.Plano).filter(models.Plano.id == transacao.plano_id).first()
                    
                    if usuario and plano:
                        usuario.status = 'ativo'
                        duracao = transacao.custom_duration_minutes if transacao.custom_duration_minutes else plano.duracao_minutos
                        usuario.fim_acesso = datetime.now() + timedelta(minutes=duracao)
                        
                    db.add(models.Log(
                        mac_address=transacao.mac_address,
                        event_type="AUTH_SUCCESS",
                        description=f"PIX Paid via WEBHOOK (IPN). Internet liberated until {usuario.fim_acesso}"
                    ))
                    db.commit()
                    print(f"🔓 [WEBHOOK OpenNDS] Internet Liberated for MAC {transacao.mac_address} until {usuario.fim_acesso}")
                    
        except Exception as e:
            print(f"Webhook error: {e}")
            return {"status": "error", "message": str(e)}
    return {"status": "ok"}

@router.post("/resgatar-voucher")
def resgatar_voucher(payload: Dict[str, Any] = Body(...), db: Session = Depends(database.get_db)):
    mac_address = payload.get("mac")
    codigo = payload.get("codigo")

    if not mac_address or not codigo:
        raise HTTPException(status_code=400, detail="Missing MAC or Voucher Code")

    # 1. Find Voucher
    voucher = db.query(models.Voucher).filter(models.Voucher.codigo == codigo).first()
    if not voucher:
        db.add(models.Log(mac_address=mac_address, event_type="AUTH_FAIL", description=f"Invalid voucher attempted: {codigo}"))
        db.commit()
        return {"sucesso": False, "erro": "Voucher não encontrado no sistema."}
    
    if voucher.status != "disponivel":
         db.add(models.Log(mac_address=mac_address, event_type="AUTH_FAIL", description=f"Redeemed or invalid voucher attempted: {codigo}"))
         db.commit()
         return {"sucesso": False, "erro": f"Voucher já resgatado ou inválido (Status: {voucher.status})."}
         
    if voucher.is_universal:
        # Check if MAC already used this universal voucher
        usage = db.query(models.VoucherUsage).filter(
            models.VoucherUsage.voucher_id == voucher.id,
            models.VoucherUsage.mac_address == mac_address
        ).first()
        if usage:
             db.add(models.Log(mac_address=mac_address, event_type="AUTH_FAIL", description=f"MAC already used universal voucher: {codigo}"))
             db.commit()
             return {"sucesso": False, "erro": "Você já resgatou este código promocional."}
             
        if voucher.uses_count >= voucher.max_uses:
             voucher.status = "esgotado"
             db.commit()
             return {"sucesso": False, "erro": "Este código promocional atingiu o limite de usos."}
    
    # 2. Find or Create User
    usuario = db.query(models.Usuario).filter(models.Usuario.mac_address == mac_address).first()
    if not usuario:
        usuario = models.Usuario(mac_address=mac_address, status='ativo')
        db.add(usuario)
        db.flush() # obtain ID without committing yet

    # 3. Burn voucher and give access
    if voucher.is_universal:
        voucher.uses_count += 1
        if voucher.uses_count >= voucher.max_uses:
            voucher.status = "esgotado"
        nova_utilizacao = models.VoucherUsage(voucher_id=voucher.id, mac_address=mac_address)
        db.add(nova_utilizacao)
    else:
        voucher.status = "usado"
        voucher.mac_address_usado = mac_address
        voucher.data_uso = datetime.now()

    usuario.status = "ativo"
    
    # Calculate expiration (additive if already active, or from now if expired)
    now = datetime.now()
    if usuario.fim_acesso and usuario.fim_acesso > now:
        usuario.fim_acesso = usuario.fim_acesso + timedelta(minutes=voucher.duracao_minutos)
    else:
        usuario.fim_acesso = now + timedelta(minutes=voucher.duracao_minutos)
    
    db.add(models.Log(mac_address=mac_address, event_type="AUTH_SUCCESS", description=f"Voucher {codigo} redeemed for {voucher.duracao_minutos} minutes."))
    db.commit()

    print(f"🔓 [OpenNDS SIMULATOR] Internet Fully Liberated for MAC {mac_address} via VOUCHER. Expires: {usuario.fim_acesso}")

    return {
        "sucesso": True,
        "plano": {
            "duracao": f"{voucher.duracao_minutos // 60} Horas",
            "horas": voucher.duracao_minutos / 60
        }
    }


