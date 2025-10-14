# Bank of Canada

endpoint: `https://www.bankofcanada.ca/valet/observations/`
base: CAD
updates: daily 4:30 PM ET
format: JSON

series pattern: `FX{CURRENCY}CAD` (e.g., FXUSDCAD, FXEURCAD)

**rate inversion**: API returns foreign-to-CAD (1 USD = 1.4 CAD)
we invert for CAD base queries (1 CAD = 0.714 USD)

23 currencies actively updated (verified oct 2025):

- excluded: MYR, THB, VND (discontinued 2019)
- includes: USD, EUR, GBP, JPY, CHF, CNY, AUD, INR, BRL, MXN, etc
