from srs_reference import *
ev=[(0,True,'M'),(0,True,'F'),(3,True,'H'),(0,False,'H'),(0,True,'M'),(0,True,'M')]
# FSRS follows its own due dates
S=D=None; day=0; due=0; print('FSRS-6 (retencao 90%)')
for late,ok,dif in ev:
    day=due+late if S is not None else 0
    g=1 if not ok else {'H':2,'M':3,'F':4}[dif]
    if S is None: S=W[g-1]; D=d0(g); r=None
    else:
        r=R(day-tl,S); S=Sfail(D,S,r) if g==1 else Ssucc(D,S,r,g); D=nextD(D,g)
    tl=day; iv=max(1,round(I(0.9,S))); due=day+iv
    print(f' dia {day:3d} nota={g} R={None if r is None else round(r,2)} S={S:.1f} D={D:.1f} proxima em {iv}d (dia {due})')
S=None; day=0; due=0; print('Simples (Hard x1.2, Medio x2, Facil x3; lapso x0.25; credito 50% do atraso)')
for late,ok,dif in ev:
    day=due+late if S is not None else 0
    if S is None:
        S=1 if not ok else {'H':1,'M':2,'F':4}[dif]; r=None
    else:
        r=0.9**((day-tl)/S)
        S=max(1,round(0.25*S,2)) if not ok else (S+0.5*late)*{'H':1.2,'M':2.0,'F':3.0}[dif]
    tl=day; iv=max(1,round(S)); due=day+iv
    print(f' dia {day:3d} {"ok  " if ok else "ERRO"}{dif} R={None if r is None else round(r,2)} S={S:.1f} proxima em {iv}d (dia {due}) dominio={min(1,math.log(1+S)/math.log(181)):.0%}')
