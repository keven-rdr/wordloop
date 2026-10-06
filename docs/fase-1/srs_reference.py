import math
W=[0.212,1.2931,2.3065,8.2956,6.4133,0.8334,3.0194,0.001,1.8722,0.1666,0.796,1.4835,0.0614,0.2629,1.6483,0.6014,1.8729,0.5425,0.0912,0.0658,0.1542]
DEC=-W[20]; FAC=0.9**(1/DEC)-1
def R(t,S): return (1+FAC*t/S)**DEC
def I(r,S): return S/FAC*(r**(1/DEC)-1)
def d0(g): return min(10,max(1,W[4]-math.exp(W[5]*(g-1))+1))
def nextD(D,g):
    D1=D-W[6]*(g-3); D1=D+(D1-D)*(10-D)/9
    return min(10,max(1,W[7]*d0(4)+(1-W[7])*D1))
def Ssucc(D,S,r,g):
    hp=W[15] if g==2 else 1; eb=W[16] if g==4 else 1
    return S*(1+math.exp(W[8])*(11-D)*S**-W[9]*(math.exp(W[10]*(1-r))-1)*hp*eb)
def Sfail(D,S,r):
    return min(S, W[11]*D**-W[12]*((S+1)**W[13]-1)*math.exp(W[14]*(1-r)))
# grades: 1 Again 2 Hard 3 Good 4 Easy ; difficulty labels H/M/F ; ok bool
def fsrs_run(seq):
    S=D=None; t_last=0; out=[]
    for day,ok,dif in seq:
        g=1 if not ok else {'H':2,'M':3,'F':4}[dif]
        if S is None: S=W[g-1]; D=d0(g); r=None
        else:
            r=R(day-t_last,S)
            if g==1: S=Sfail(D,S,r)
            else: S=Ssucc(D,S,r,g)
            D=nextD(D,g)
        S=max(S,0.01); t_last=day
        iv=max(1,round(I(0.9,S)))
        out.append((day,g,None if r is None else round(r,3),round(S,2),round(D,2),iv))
    return out
CAP=365
def simple_run(seq,due=None):
    S=None; t_last=0; out=[]; due_day=0
    for day,ok,dif in seq:
        late=max(0,day-due_day) if S is not None else 0
        if S is None:
            S=1 if not ok else {'H':1,'M':2,'F':4}[dif]; r=None
        else:
            r=0.9**((day-t_last)/S)
            if not ok: S=max(1,round(0.25*S,2))
            else:
                S=(S+0.5*late)*{'H':1.2,'M':2.0,'F':3.0}[dif]
        S=min(CAP,S); t_last=day; due_day=day+max(1,round(S))
        dom=min(1,math.log(1+S)/math.log(181))
        out.append((day,'ok' if ok else 'ERRO',dif,None if r is None else round(r,3),round(S,2),max(1,round(S)),round(dom,2)))
    return out
if __name__=='__main__':
    print('Good-only FSRS-6 check (doc table):')
    S=W[2]; S=S*math.exp(W[17]*(0+W[18]))*S**-W[19]; print(' same-day S',round(S,3))
    D=d0(3); t=0
    for i in range(5):
        iv=round(I(0.9,S)); r=R(iv,S); S=Ssucc(D,S,r,3); D=nextD(D,3); print(' review',i+3,'interval',iv,'-> S',round(S,1))
    # scenario
    seq=[(0,True,'M'),(2,True,'F'),(9,True,'H'),(19,False,'H'),(20,True,'M'),(23,True,'M')]
    print('\nFSRS-6 scenario'); [print(' ',x) for x in fsrs_run(seq)]
    print('\nSimple scenario'); [print(' ',x) for x in simple_run(seq)]
