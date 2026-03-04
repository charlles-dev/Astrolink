import re
import os

html_path = 'index.html'
with open(html_path, 'r', encoding='utf-8') as f:
    html = f.read()

config_script = """
    <script>
        tailwind.config = {
            darkMode: 'class',
            theme: {
                extend: {}
            }
        }
    </script>
"""
if "tailwind.config" not in html:
    html = html.replace('</head>', config_script + '</head>')

# Add toggle button to sidebar
toggle_html = """
            <nav class="flex-1 px-4 py-6 space-y-2">
                <button onclick="toggleTheme()" id="btn-theme-toggle"
                    class="w-full flex items-center justify-between px-4 py-3 text-slate-500 hover:bg-slate-200 dark:text-slate-400 dark:hover:bg-slate-800/50 dark:hover:text-cyan-300 rounded-xl transition-colors font-medium mb-4 shadow-sm border border-slate-200 dark:border-white/5">
                    <span class="flex items-center gap-3"><i class="fas fa-moon w-5"></i> Tema Escuro/Claro</span>
                </button>
"""
html = html.replace('<nav class="flex-1 px-4 py-6 space-y-2">', toggle_html)

# Add class variable toggles
def replace_class(old, new_light, old_dark=None):
    if old_dark is None:
        old_dark = f"dark:{old}"
    global html
    # Match the exact class name
    html = re.sub(rf'(?<!-)\b{re.escape(old)}\b', f'{new_light} {old_dark}', html)

# Backgrounds
replace_class('bg-[#030712]', 'bg-slate-50')
replace_class('bg-slate-900/40', 'bg-white/70')
replace_class('bg-slate-950/80', 'bg-white')
replace_class('bg-slate-900/50', 'bg-white/80')
replace_class('bg-slate-800', 'bg-slate-200')
replace_class('bg-slate-950/50', 'bg-slate-100')
replace_class('bg-slate-800/50', 'bg-slate-200/50')
replace_class('bg-slate-900/80', 'bg-white/90')
replace_class('bg-slate-900', 'bg-white')
replace_class('bg-[#0f172a]', 'bg-slate-100')

# Text Colors
replace_class('text-slate-50', 'text-slate-900')
replace_class('text-white', 'text-slate-900')
replace_class('text-slate-300', 'text-slate-700')
replace_class('text-slate-400', 'text-slate-600')
replace_class('text-slate-500', 'text-slate-500')

# Borders
replace_class('border-white/10', 'border-slate-300/50')
replace_class('border-white/5', 'border-slate-200/60')
replace_class('border-slate-700/50', 'border-slate-300')
replace_class('border-slate-700', 'border-slate-300')

print("Writing modified HTML")
with open(html_path, 'w', encoding='utf-8') as f:
    f.write(html)
