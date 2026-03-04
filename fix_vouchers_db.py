import sqlite3
import os

db_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'db', 'hotspot.db')
print(f"Connecting to DB at: {db_path}")

try:
    conn = sqlite3.connect(db_path)
    c = conn.cursor()
    
    try:
        c.execute('ALTER TABLE vouchers ADD COLUMN nome VARCHAR(50)')
        print("Column 'nome' added.")
    except Exception as e:
        print(f"Error adding 'nome': {e}")
        
    try:
        c.execute('ALTER TABLE vouchers ADD COLUMN is_universal BOOLEAN NOT NULL DEFAULT 0')
        print("Column 'is_universal' added.")
    except Exception as e:
        print(f"Error adding 'is_universal': {e}")

    try:
        c.execute('ALTER TABLE vouchers ADD COLUMN max_uses INTEGER NOT NULL DEFAULT 1')
        print("Column 'max_uses' added.")
    except Exception as e:
        print(f"Error adding 'max_uses': {e}")

    try:
        c.execute('ALTER TABLE vouchers ADD COLUMN uses_count INTEGER NOT NULL DEFAULT 0')
        print("Column 'uses_count' added.")
    except Exception as e:
        print(f"Error adding 'uses_count': {e}")

    conn.commit()
    print("Database update complete!")
except Exception as e:
    print(f"Fatal error: {e}")
finally:
    if 'conn' in locals():
        conn.close()
