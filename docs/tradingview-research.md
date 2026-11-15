# TradingView chart-tool research

วันที่ค้นคว้า: 15 พฤศจิกายน 2026

## ข้อสรุปและขอบเขตหลักฐาน

เลือกเครื่องมือตามหน้าที่และวัดผลหลังต้นทุน ไม่พบหลักฐานจากแหล่งที่ตรวจว่า indicator ตัวเดียวแม่นที่สุดในทุกตลาดหรือทุก timeframe ความถูกต้องของสูตร การบอกแนวโน้มปัจจุบัน และผลกำไรในอนาคตเป็นคนละเรื่อง

TradingView เรียกองค์ประกอบของ Technical Ratings ว่า popular indicators แต่ไม่ได้ให้สถิติ usage ที่ใช้จัดอันดับ built-in ทุกตัวได้ จึงไม่อ้างอันดับความนิยมจากจำนวน likes ของสคริปต์ชุมชน ตารางนี้เลือกเครื่องมือมาตรฐานจากเอกสารโดยตรง ไม่ใช่ leaderboard ความแม่นยำ [Technical Ratings](https://www.tradingview.com/support/solutions/43000614331-technical-ratings/)

งาน Deprez–Frömmel ศึกษากฎ 75,360 แบบ โดยคิด transaction costs, ทดสอบ out-of-sample และควบคุม data mining ผลต่างกันตามช่วงเวลาและความถี่ บางกลุ่มช่วยด้าน risk-adjusted performance ได้ แต่ไม่ได้พิสูจน์ MACD/Supertrend หรือ parameter ชุดใดเป็นผู้ชนะถาวร และข้อมูล Bitstamp BTC ไม่ใช่หลักฐานตรงสำหรับ Bitkub THB [บทความฉบับผู้เขียน](https://backoffice.biblio.ugent.be/download/01HY3C3S169G1N6QNYR55NZMFB/01HY60XZGZYHNQ6188MSVJT0SG)

## เครื่องมือและการตัดสินใจ

| กลุ่ม | สิ่งที่อ่านได้ | สิ่งที่ต้องระวัง | การเปลี่ยนในโปรเจกต์ |
|---|---|---|---|
| EMA / SMA | ระดับและทิศทางค่าเฉลี่ยราคา | ตามหลังราคา; seed และความยาวประวัติมีผล | เก็บ EMA เดิม; native trend ขอ 1000 แท่ง ลดอิทธิพล seed; เพิ่ม EMA20/50 crossover ให้ backtest |
| RSI14 / ROC | momentum และการเปลี่ยนราคา | overbought ไม่ได้แปลว่าราคาต้องลง | เก็บ; ไม่มีการเพิ่มเสียงโหวตซื้ออิสระจากข้อมูลราคาเดียวกัน |
| DMI14 / ADX14 | ADX วัดความแรง; +DI/-DI วัดทิศทาง | ADX สูงไม่ได้แปลว่าขาขึ้น; เกณฑ์ไม่ใช่กฎสากล | OHLC regime ใช้ ADX ≥25, DI และทิศทาง net price ที่ตรงกัน; ลบทางลัดนับแท่งที่เคยข้าม ADX |
| MACD12/26/9 | ความต่าง EMA และการเปลี่ยน momentum | cross หลอกในช่วงแกว่ง | เพิ่มใน `analyze_chart`; เพิ่ม `macd_cross` ใน backtest |
| Bollinger20/2 | ช่วงราคาเทียบ SMA และความผันผวน | แตะ band ไม่ได้เป็นคำสั่งซื้อ/ขาย; ราคาเดินตาม band ได้ | เพิ่ม basis/bands, bandwidth, %B และ cross; เพิ่ม `bollinger_breakout` ใน backtest |
| Supertrend10/3 | เส้นแนวโน้มตาม ATR และการ flip | แพ้การแกว่งกลับไปมา; เปลี่ยน multiplier แล้วผลเปลี่ยน | เพิ่มค่าและ flip บนราคาปิด; wick อย่างเดียวไม่ flip; เพิ่ม `supertrend_flip` ใน backtest |
| Anchored VWAP | HLC3 ถ่วงน้ำหนักด้วย volume ตั้งแต่ anchor | anchor/session/หน่วย volume ต้องตรง; ไม่ใช่ tick VWAP | เพิ่ม; default UTC midnight; ถ้าข้อมูลไม่ครอบคลุมหรือ anchor ไม่ตรง candle-open จะคืน unavailable |
| ATR14 | ความผันผวนและระยะ stop | ไม่บอกทิศทาง; แต่ละ timeframe ต่างกัน | เก็บ daily ATR gate และแยก entry/trend ATR เดิม |
| Breakout / Pullback | จุดเข้าแบบกฎที่ตรวจได้ | ต้องมี trend, volume, liquidity และต้นทุนร่วม | เก็บและนำมาเป็น baseline เปรียบเทียบ; ไม่รับรองจากสัญญาณล้วน |
| CAPM | ผลตอบแทนตามสมมติฐาน Rf, beta, market return | ไม่ได้อ่านแท่งเทียนหรือ fit การคาดการณ์ราคา | เก็บ API เดิมเพื่อไม่ทำลาย callers; ย้ายออกจากหมวด chart indicators และระบุข้อจำกัดใน tool description |
| Volume Profile / Footprint | volume ตามระดับราคา | ต้องมีข้อมูลละเอียดกว่า OHLCV ที่ส่งมา | ยังไม่เพิ่ม ไม่กระจาย volume ลงราคาแล้วอ้างว่าเทียบเท่า TradingView |
| Technical Ratings / คะแนนโหวตรวม | สรุปเงื่อนไขจากหลาย indicator | คะแนนไม่ใช่ win probability; indicators หลายตัวสัมพันธ์กัน | ไม่เพิ่มเป็น gate ซื้อจริง; ยกเลิกการตีความ heuristic score เป็นความมั่นใจเชิงสถิติ |

สูตรและข้อจำกัดอ้างอิงจาก [Moving Averages](https://www.tradingview.com/support/solutions/43000502589-moving-averages/), [DMI](https://www.tradingview.com/support/solutions/43000502250-directional-movement-dmi/), [MACD](https://www.tradingview.com/support/solutions/43000502344-moving-average-convergence-divergence-macd-indicator/), [Bollinger Bands](https://www.tradingview.com/support/solutions/43000501840-bollinger-bands-bb/), [Supertrend](https://www.tradingview.com/support/solutions/43000634738-supertrend/), [VWAP](https://www.tradingview.com/support/solutions/43000502018-volume-weighted-average-price-vwap/), [ATR](https://www.tradingview.com/support/solutions/43000501823-average-true-range-atr/) และ [Volume Profile methodology](https://www.tradingview.com/support/solutions/43000502040-volume-profile-indicators-basic-concepts/)

## สูตรที่ใช้และความต่างจากหน้าจอ TradingView

| ค่า | นิยามในโปรเจกต์ |
|---|---|
| EMA | SMA seed ที่ period แรก จากนั้น alpha=2/(period+1); รักษา API เดิม |
| MACD | EMA12 − EMA26; signal=EMA9 ของ MACD ที่ seed แล้ว; histogram=MACD−signal |
| Bollinger | SMA20 ของ close ±2×population standard deviation; %B เป็น null เมื่อ bands แบน |
| Supertrend | HLC midpoint ±3×Wilder ATR10 พร้อม upper/lower bands ที่เลื่อนตามกฎ; เริ่ม down จน ATR พร้อม; flip ด้วย close |
| DMI | DM/TR ใช้ Wilder smoothing; ADX เฉลี่ย DX แล้ว smooth ต่อ; ส่ง +DI/-DI แยกจาก ADX |
| VWAP | sum(HLC3×volume)/sum(volume) ตั้งแต่ candle-open ที่ระบุ; synthetic/zero-volume ไม่เพิ่มน้ำหนัก |

เราไม่ได้อ้าง numerical parity แบบทุกหลักกับ built-ins ของ TradingView: EMA/MACD ใช้ seed เดิมของ repo และ ATR เริ่มจากแท่งที่มี previous close แล้ว ต่างจาก Pine ที่ใช้ high−low ได้เมื่อ previous close ไม่มี ต้องเทียบ feed, history start, timezone, anchor และ smoothing เดียวกันก่อนสรุปว่าสูตรผิด เก็บความละเอียดตัวเลขไว้จนถึง exchange rounding [Pine built-ins](https://www.tradingview.com/pine-script-docs/language/built-ins/)

ใช้เฉพาะแท่งปิด ไม่ใช้ future pivots หรือข้อมูล higher-timeframe ที่ยังไม่ปิด เพราะค่าบนแท่ง realtime เปลี่ยนได้ และการวาดย้อนหลังทำให้ดูแม่นกว่าที่รู้ได้จริงขณะนั้น [TradingView: Repainting](https://www.tradingview.com/pine-script-docs/concepts/repainting/)

## หลักฐานจากข้อมูลสาธารณะ

ไฟล์ [snapshot และผลเต็ม](research/tradingview-evaluation.json) เก็บ candles, metadata, fixed cutoff, ค่าใช้จ่าย, trades, per-fold metrics และ fingerprint ของ source ที่ประเมิน

- Cutoff: 2026-11-15 04:12:15 UTC / 11:12:15 Asia/Bangkok
- BTC_THB และ ETH_THB อย่างละ 1000 แท่งต่อ timeframe: 15m และ 4H
- 15m ครอบคลุม 4–15 พ.ย. 2026; 4H ครอบคลุม 1 มิ.ย.–14 พ.ย. 2026; 220 แท่งแรกเป็น warmup ไม่ใช่ช่วงซื้อขายที่วัดผล
- 6 กลยุทธ์ × 4 datasets × 2 cost assumptions = 48 evaluations; ไม่มีการ fit parameter หรือเลือกผู้ชนะแล้วนำไปวัดบนข้อมูลเดิมว่าเป็น OOS
- ค่าธรรมเนียม 0.25% ต่อขา; slippage 0.1% และ stress 0.2% ต่อขา; target distance 2R; 3 chronological folds
- Source เป็น public MCP market-history tool; ไม่เรียก wallet, fee account, open orders หรือส่งออเดอร์

| กรณี cost พื้นฐาน | กลยุทธ์ | เทรด | ชนะ | ผลสุทธิเฉลี่ย/เทรด | Profit factor แบบ notional เท่ากัน |
|---|---|---:|---:|---:|---:|
| BTC 15m | breakout | 7 | 0.00% | −1.0467% | 0.0000 |
| BTC 4H | breakout | 16 | 37.50% | −0.4470% | 0.7074 |
| BTC 4H | MACD cross | 28 | 17.86% | −1.3915% | 0.2171 |
| ETH 4H | pullback | 16 | 37.50% | −0.4566% | 0.7626 |
| ETH 4H | Supertrend flip | 9 | 22.22% | −1.4189% | 0.3158 |
| ETH 4H | Bollinger breakout | 23 | 17.39% | −1.6567% | 0.2582 |
| ETH 4H | EMA20/50 crossover | 4 | 50.00% | +0.1441% | 1.1041 |

46 จาก 48 evaluations มีผลสุทธิเฉลี่ยติดลบ ตัวที่บวกทั้งสอง cost assumptions คือ ETH 4H EMA20/50 แต่มีเพียง 4 เทรด ช่วง Wilson 95% ของ win rate ประมาณ 15–85% และยังไม่ปรับ selection bias หรือความสัมพันธ์ของ trades จึงไม่รองรับคำกล่าวว่าแม่นที่สุด ผลนี้เป็น entry-only experiment ไม่ใช่ผลของ native multi-symbol workflow ที่มี regime/EMA/RS/daily ATR/volume/risk gates

ผลวิจัยจึงไม่รองรับการเปลี่ยนค่าเริ่มต้นไปซื้อเมื่อ MACD/Supertrend/Bollinger ส่งสัญญาณล้วน และไม่รองรับการลบ indicator เพียงเพราะตัวอย่างกลยุทธ์หนึ่งขาดทุน คงเครื่องมือคำนวณไว้สำหรับตรวจสภาวะตลาด; native analyzer แนบ chart diagnostics โดยไม่ยกเลิก validation gate หรือเพิ่มสัญญาณใหม่เป็นสิทธิ์ซื้ออัตโนมัติ

Backtest แก้การมองข้อมูลอนาคตผ่านการเลือก “last actual candle” ล่วงหน้าแล้ว: ถ้า fold จบด้วยแท่งไม่มีธุรกรรมและยังถืออยู่ จะรายงาน unresolved position และ `complete=false` ไม่ย้อนปิดที่แท่งก่อนหน้าเพื่อให้มีผลลัพธ์ครบ

## ทำซ้ำและตรวจสอบ

```powershell
rtk go build -buildvcs=false -o "$env:TEMP\gokub-chart-research.exe" .
rtk bun scripts/research-chart-strategies.mjs "$env:TEMP\gokub-chart-research.exe" docs/research/tradingview-evaluation.json
```

ไม่ส่ง saved snapshot หากต้องการเก็บข้อมูลสาธารณะชุดใหม่ ผลลัพธ์เขียนกลับ `docs/research/tradingview-evaluation.json` สคริปต์เริ่ม MCP ใน temporary directory ที่ไม่มี `.env` และส่ง API credentials เป็นค่าว่าง

คำสั่งตรวจ regression tests สำหรับผู้ใช้ ตามข้อกำหนด repo ที่ไม่ให้ agent รัน tests:

```powershell
rtk go test -buildvcs=false -count=1 ./...
```

หลักฐานตรวจงาน: build และ `go vet -buildvcs=false ./...` ผ่าน; MCP initialize/tool discovery พบ 42 tools และ replay จาก snapshot ผ่านครบ 48 evaluations, `errors=[]`, ทุกผล `complete=true`

หลังผู้ใช้สั่ง build/install และทดสอบเพิ่มเติม: แก้ลำดับคำนวณค่าเฉลี่ย Bollinger ให้รวมส่วนต่างก่อนหาร เพื่อลด floating-point error โดยไม่แก้ assertion; `go test -buildvcs=false -count=1 ./...` ผ่าน 186 รายการใน 6 packages ติดตั้ง binary ทับ `D:\home\.go\bin\gokub-mcp.exe` และทดสอบ MCP replay ข้างต้นผ่าน binary ที่ติดตั้งจริงแล้ว ยังไม่ได้รัน authenticated native scan และยังไม่ได้ตรวจ numerical parity กับ TradingView export จริง

ข้อจำกัดที่ยังต้องตรวจต่อก่อนอ้างความแม่นใน production: ช่วง holdout ที่ไม่ใช้เลือกกฎ, หลายสภาวะตลาดและเหรียญ, liquidity/fills จริง, portfolio sizing, intratrade drawdown และการเทียบกับ TradingView export ที่ใช้ history/initialization ตรงกัน
