# Solvencia y estrés

## Propósito

`src/solvency` convierte una posición agregada en una evaluación conservadora. No cambia saldos ni autoriza comandos. Su resultado permite definir alertas, límites y revisiones manuales antes de una ventana de liquidación.

```mermaid
flowchart LR
    S["Snapshot por activo"] --> V["Validación"]
    V --> H["Haircuts"]
    V --> X["Shocks"]
    V --> C["Concentración"]
    H --> W["Waterfall"]
    X --> W
    C --> W
    W --> B["Buffer y coverage bps"]
    B --> G["Banda y señales"]
```

## Waterfall

La liquidez disponible recibe un haircut. Impagos y claims pendientes reciben shocks con redondeo superior. Las recuperaciones se reconocen después de un haircut y la mayor facility añade un cargo de concentración.

```text
available = floor(pool × (1 - liquidityHaircut))
stressedDefaults = ceil(pendingDefaults × (1 + defaultShock))
stressedClaims = ceil(pendingClaims × (1 + claimShock))
recognizedRecoveries = floor(expectedRecoveries × (1 - recoveryHaircut))
concentrationAddon = ceil(largestFacility × concentrationAddonRate)
required = max(stressedDefaults + stressedClaims + concentrationAddon
               - recognizedRecoveries, 0)
coverageBps = floor(available × 10_000 / required)
```

```mermaid
flowchart TD
    P["Pool 10.000"] -->|haircut 10%| L["Liquidez 9.000"]
    D["Defaults 2.000"] -->|shock 25%| SD["2.500"]
    C["Claims 1.000"] -->|shock 15%| SC["1.150"]
    F["Mayor facility 5.000"] -->|add-on 12,5%| A["625"]
    R["Recoveries 500"] -->|haircut 60%| RR["200"]
    SD --> Q["Requerido 4.075"]
    SC --> Q
    A --> Q
    RR --> Q
    L --> Z["Buffer 4.925"]
    Q --> Z
```

## Bandas

| Banda      | Regla por defecto                       | Acción orientativa    |
| ---------- | --------------------------------------- | --------------------- |
| `nominal`  | cobertura ≥ 13.500 bps o requerido cero | operación normal      |
| `watch`    | cobertura ≥ 11.500 bps                  | seguimiento reforzado |
| `guarded`  | cobertura ≥ 9.000 bps                   | limitar crecimiento   |
| `critical` | cobertura < 9.000 bps                   | pausa y revisión      |

```mermaid
stateDiagram-v2
    Nominal --> Watch: menor cobertura
    Watch --> Guarded: cruza mínimo
    Guarded --> Critical: menor a 9.000 bps
    Critical --> Guarded: recapitalización
    Guarded --> Watch: exposición reducida
    Watch --> Nominal: buffer restaurado
```

Las señales adicionales indican buffer negativo, concentración superior al 50%, reserva inferior a exposición y claims estresados por encima de liquidez. Su orden es determinista para facilitar alertas y comparaciones.

## Ejecución recomendada

Calcule una posición por activo después de cada lote confirmado y antes de abrir nuevas facilidades. Mantenga la política de stress versionada, almacene input y output y compare la banda con el límite aprobado. Un cambio de parámetros debe pasar revisión independiente y análisis retrospectivo.
