import sqlite3
import os

db_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'db', 'hotspot.db')
print(f"Connecting to DB at: {db_path}")

try:
    conn = sqlite3.connect(db_path)
    c = conn.cursor()
    
    try:
        c.execute('ALTER TABLE planos ADD COLUMN max_devices INTEGER NOT NULL DEFAULT 1')
        print("Column 'max_devices' added to planos.")
    except Exception as e:
        print(f"Error adding 'max_devices': {e}")
        
    conn.commit()
    print("Database update complete!")
except Exception as e:
    print(f"Fatal error: {e}")
finally:
    if 'conn' in locals():
        conn.close()
