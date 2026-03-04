from sqlalchemy import Boolean, Column, ForeignKey, Integer, String, Float, DateTime
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from database import Base

class Usuario(Base):
    __tablename__ = "usuarios"

    id = Column(Integer, primary_key=True, index=True, autoincrement=True)
    mac_address = Column(String(17), unique=True, index=True, nullable=False)
    status = Column(String(20), default='bloqueado', nullable=False)
    fim_acesso = Column(DateTime, nullable=True)
    ip_atual = Column(String(15), nullable=True)
    criado_em = Column(DateTime, server_default=func.now())
    
    transacoes = relationship("TransacaoPix", back_populates="usuario")
    vouchers = relationship("Voucher", back_populates="usuario")

class Plano(Base):
    __tablename__ = "planos"

    id = Column(Integer, primary_key=True, index=True, autoincrement=True)
    nome = Column(String(50), nullable=False)
    preco = Column(Float, nullable=False)
    duracao_minutos = Column(Integer, nullable=False)
    max_devices = Column(Integer, default=1, nullable=False)
    ativo = Column(Boolean, default=True)

    transacoes = relationship("TransacaoPix", back_populates="plano")

class TransacaoPix(Base):
    __tablename__ = "transacoes_pix"

    id = Column(Integer, primary_key=True, index=True, autoincrement=True)
    txid = Column(String(100), unique=True, index=True, nullable=False)
    mac_address = Column(String(17), ForeignKey("usuarios.mac_address"), nullable=False)
    plano_id = Column(Integer, ForeignKey("planos.id"), nullable=False)
    status = Column(String(20), default='pendente', nullable=False)
    custom_duration_minutes = Column(Integer, nullable=True)
    data_criacao = Column(DateTime, server_default=func.now())

    usuario = relationship("Usuario", back_populates="transacoes")
    plano = relationship("Plano", back_populates="transacoes")

class Voucher(Base):
    __tablename__ = "vouchers"

    id = Column(Integer, primary_key=True, index=True, autoincrement=True)
    codigo = Column(String(20), unique=True, index=True, nullable=False)
    nome = Column(String(50), nullable=True) # Ex: "Promoção Dia das Mães"
    duracao_minutos = Column(Integer, nullable=False)
    valor_pago = Column(Float, nullable=False, default=0.0)
    status = Column(String(20), default='disponivel', nullable=False)
    
    # Generic usage settings
    is_universal = Column(Boolean, default=False, nullable=False)
    max_uses = Column(Integer, default=1, nullable=False)
    uses_count = Column(Integer, default=0, nullable=False)
    
    # For single-use fallback (legacy)
    mac_address_usado = Column(String(17), ForeignKey("usuarios.mac_address"), nullable=True)
    data_criacao = Column(DateTime, server_default=func.now())
    data_uso = Column(DateTime, nullable=True)

    usuario = relationship("Usuario", back_populates="vouchers")
    usages = relationship("VoucherUsage", back_populates="voucher")

class VoucherUsage(Base):
    """ Record every time a MAC address uses a voucher """
    __tablename__ = "voucher_usages"

    id = Column(Integer, primary_key=True, index=True, autoincrement=True)
    voucher_id = Column(Integer, ForeignKey("vouchers.id"), nullable=False)
    mac_address = Column(String(17), ForeignKey("usuarios.mac_address"), nullable=False)
    data_uso = Column(DateTime, server_default=func.now())

    voucher = relationship("Voucher", back_populates="usages")

class Log(Base):
    __tablename__ = "logs"

    id = Column(Integer, primary_key=True, index=True, autoincrement=True)
    mac_address = Column(String(17), nullable=True, index=True)
    event_type = Column(String(50), nullable=False, index=True) # AUTH_SUCCESS, AUTH_FAIL, PIX_CREATED, VOUCHER_GENERATED
    description = Column(String(255), nullable=True)
    created_at = Column(DateTime, server_default=func.now())

class WalledGardenRule(Base):
    __tablename__ = "walled_garden_rules"

    id = Column(Integer, primary_key=True, index=True, autoincrement=True)
    tipo = Column(String(10), nullable=False) # IP, URL, DOMAIN
    valor = Column(String(255), nullable=False) # e.g. 192.168.1.1 or *.mercadopago.com
    descricao = Column(String(255), nullable=True)
    ativo = Column(Boolean, default=True)
    created_at = Column(DateTime, server_default=func.now())

class SystemSettings(Base):
    __tablename__ = "system_settings"

    key = Column(String(50), primary_key=True, index=True)
    value = Column(String(255), nullable=True)
    description = Column(String(255), nullable=True)
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())

class ContaPersistente(Base):
    """ Record for fixed/monthly accounts that bypass standard expiration """
    __tablename__ = "contas_persistentes"

    id = Column(Integer, primary_key=True, index=True, autoincrement=True)
    mac_address = Column(String(17), ForeignKey("usuarios.mac_address"), unique=True, nullable=False)
    nome = Column(String(100), nullable=False)
    telefone = Column(String(20), nullable=True)
    email = Column(String(100), nullable=True)
    observacoes = Column(String(255), nullable=True)
    data_renovacao = Column(DateTime, nullable=True)
    ativo = Column(Boolean, default=True)
    criado_em = Column(DateTime, server_default=func.now())

    usuario = relationship("Usuario", backref="conta_persistente")

class BlacklistMAC(Base):
    __tablename__ = "blacklist_macs"

    id = Column(Integer, primary_key=True, index=True, autoincrement=True)
    mac_address = Column(String(17), unique=True, index=True, nullable=False)
    motivo = Column(String(255), nullable=True)
    data_bloqueio = Column(DateTime, server_default=func.now())
