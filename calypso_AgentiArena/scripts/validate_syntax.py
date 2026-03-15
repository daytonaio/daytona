import ast, glob, os
for d in sorted(glob.glob('*-agent')):
    f = os.path.join(d, 'main.py')
    if not os.path.exists(f): continue
    try:
        ast.parse(open(f, encoding='utf-8').read())
        print(f"  OK: {f}")
    except SyntaxError as e:
        print(f"  BROKEN: {f} -> {e}")
