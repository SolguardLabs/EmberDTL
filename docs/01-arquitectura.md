# Arquitectura de EmberDTL

## Objetivo

EmberDTL separa la decisión económica, la transición contable y la presentación de resultados. El núcleo Go no consulta red, reloj de sistema ni bases de datos: recibe una política, un estado inicial y una secuencia ordenada. Esta frontera hace que una misma entrada produzca el mismo informe.

```mermaid
flowchart TB
    subgraph Entrada
        JSON["Scenario JSON"]
        SDK["TypeScript SDK"]
    end
    subgraph Núcleo
        SC["scenario"]
        EN["engine"]
        PO["policy"]
        RI["risk"]
        IN["insurance"]
        LE["ledger"]
        DO["domain"]
    end
    subgraph Salida
        RE["SystemReport"]
        RC["Reconciliation"]
        EV["Events"]
    end
    JSON --> SC
    SDK --> JSON
    SC --> EN
    EN --> PO
    EN --> RI
    EN --> IN
    EN --> LE
    PO --> DO
    RI --> DO
    IN --> DO
    LE --> DO
    DO --> RE
    RE --> RC
    RE --> EV
```

## Agregados

`State` contiene activos, cuentas, pools, reservas, facilidades, impagos, reclamaciones, eventos y métricas. Cada mapa tiene una clave estable; las proyecciones se ordenan antes de serializarse para evitar diferencias por iteración de mapas.

```mermaid
erDiagram
    ASSET ||--|| INSURANCE_POOL : denomina
    ASSET ||--o{ RESERVE_BUCKET : segrega
    ACCOUNT ||--o{ RESERVE_BUCKET : posee
    RESERVE_BUCKET ||--o{ FACILITY : financia
    FACILITY ||--o| DEFAULT_CASE : origina
    DEFAULT_CASE ||--o{ CLAIM : agrupa
    ACCOUNT ||--o{ CLAIM : cobra
    FACILITY ||--o{ EVENT : registra
```

Una `Facility` conserva principal, outstanding, repaid, fees, época de apertura y vencimiento. Un `DefaultCase` conserva importe reportado, techo de cobertura, exposición pendiente, cobertura solicitada y recuperación. El `InsurancePool` concentra contribuciones, comisiones, pagos y recuperaciones de un solo activo.

## Flujo de comando

```mermaid
sequenceDiagram
    participant C as Caller
    participant S as Scenario
    participant E as Engine
    participant D as Domain
    participant J as Journal
    C->>S: step
    S->>E: request tipada
    E->>E: normaliza IDs y política
    E->>D: comprueba precondiciones
    E->>D: aplica transición
    E->>J: registra asiento y evento
    E-->>S: resultado
    S-->>C: informe ordenado
```

Los errores se clasifican en entrada inválida, entidad ausente, política, riesgo, estado o solvencia. El caller debe conservar la categoría y no reintentar automáticamente errores deterministas.

## Fronteras de extensión

- Nuevas políticas: funciones puras en `src/policy`.
- Nuevas métricas de stress: `src/solvency`, sin mutar el estado.
- Nuevos comandos: request en `engine`, caso en `scenario` y evento asociado.
- Nuevas proyecciones: tipos de reporte ordenados y compatibles hacia atrás.
- Clientes: procesos sin shell y decodificación estricta.

Una extensión debe mantener compatibilidad del JSON o introducir una versión explícita de esquema. Los campos nuevos deben ser opcionales para consumidores anteriores cuando no alteren la interpretación económica.
