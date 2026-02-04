import React, { useEffect, useState, useRef } from 'react';
import { Centrifuge } from 'centrifuge';

const TradingMonitor = () => {
    // --- STATE TANIMLARI ---
    // Fiyat
    const [price, setPrice] = useState('---');
    // Cüzdan (Başlangıçta 0 gösteriyoruz, veri gelince güncellenecek)
    const [wallet, setWallet] = useState({ usdt: 0.00, btc: 0.00000 });
    // Bağlantı Durumu
    const [status, setStatus] = useState('Bağlanıyor...');
    const [isConnected, setIsConnected] = useState(false);
    // Loglar
    const [logs, setLogs] = useState([]);

    const centrifugeRef = useRef(null);

    // --- YARDIMCI FONKSİYONLAR ---
    // Log Ekleme (Son 50 kaydı tutar)
    const addLog = (msg, type = 'info') => {
        const time = new Date().toLocaleTimeString();
        setLogs(prevLogs => [{ id: Date.now(), time, msg, type }, ...prevLogs].slice(0, 50));
    };

    // Log Renkleri (Tailwind)
    const getLogClass = (type) => {
        switch (type) {
            case 'success': return 'text-green-400';
            case 'error': return 'text-red-500';
            case 'warning': return 'text-yellow-400 font-bold'; // Sinyaller için
            case 'buy': return 'text-green-500 font-bold uppercase'; // Alım işlemi
            case 'sell': return 'text-red-500 font-bold uppercase'; // Satım işlemi
            default: return 'text-gray-400';
        }
    };

    // --- WEBSOCKET BAĞLANTISI ---
    useEffect(() => {
        // DİKKAT: Port 8085 (Backend'de ayarladığımız port)
        const cent = new Centrifuge('ws://localhost:8085/connection/websocket', { 
            debug: false // Konsolu kirletmemesi için kapattım, hata ararken true yapabilirsin
        });
        
        centrifugeRef.current = cent;

        // 1. Bağlantı Olayları
        cent.on('connected', (ctx) => {
            setStatus(`BAĞLANDI (Client ID: ${ctx.client})`);
            setIsConnected(true);
            addLog('✅ Sunucu bağlantısı başarılı.', 'success');
        });

        cent.on('disconnected', (ctx) => {
            setStatus(`KOPTU: ${ctx.reason}`);
            setIsConnected(false);
            addLog(`❌ Bağlantı koptu: ${ctx.reason}`, 'error');
        });

        // 2. KANAL: KLINE (Fiyat Verisi)
        const subKline = cent.newSubscription('kline');
        subKline.on('publication', (ctx) => {
            const data = ctx.data;
            if (data.close) {
                setPrice(parseFloat(data.close).toFixed(2));
            }
        });
        subKline.subscribe();

        // 3. KANAL: SIGNALS (Al-Sat Sinyalleri)
        const subSignals = cent.newSubscription('signals');
        subSignals.on('publication', (ctx) => {
            const signal = ctx.data;
                // YENİ HALİ (BUNU YAPIŞTIR):
            // Go'dan gelen JSON muhtemelen küçük harfli (action, reason, symbol)
            const action = signal.action || signal.Action; // Her ihtimale karşı ikisini de dene
            const reason = signal.reason || signal.Reason;
            const symbol = signal.symbol || signal.Symbol;

            const logType = action === 'BUY' || action === 'buy' ? 'buy' : 'sell';

            addLog(`🚀 SİNYAL GELDİ: ${action} - ${reason} (${symbol})`, logType);
        });
        subSignals.subscribe();

        // 4. KANAL: WALLET (Cüzdan Güncellemeleri)
        const subWallet = cent.newSubscription('wallet');
        subWallet.on('publication', (ctx) => {
            const data = ctx.data;
            setWallet({ 
                usdt: parseFloat(data.usdt), 
                btc: parseFloat(data.btc) 
            });
            // Hangi bakiyenin değiştiğini loglayalım
            addLog(`💰 Cüzdan Güncellendi: ${parseFloat(data.usdt).toFixed(2)}$ / ${parseFloat(data.btc).toFixed(5)} BTC`, 'success');
        });
        subWallet.subscribe();

        // Bağlan
        cent.connect();

        // Cleanup (Sayfadan çıkınca bağlantıyı kes)
        return () => {
            if (centrifugeRef.current) cent.disconnect();
        };
    }, []);

    // --- RENDER (Görünüm) ---
    return (
        <div className="min-h-screen bg-[#0d0d0d] text-gray-200 font-mono p-4 md:p-8 flex flex-col items-center selection:bg-green-900 selection:text-white">
            
            {/* ÜST BAŞLIK */}
            <div className="w-full max-w-4xl flex flex-col md:flex-row justify-between items-center border-b border-gray-800 pb-4 mb-8 gap-4">
                <h1 className="text-3xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-green-400 to-emerald-600 tracking-wider">
                    🦅 V2 TRADING BOT
                </h1>
                
                {/* Durum Işığı */}
                <div className={`flex items-center gap-2 px-4 py-1 rounded-full border ${isConnected ? 'border-green-900 bg-green-900/20 text-green-400' : 'border-red-900 bg-red-900/20 text-red-500'}`}>
                    <span className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`}></span>
                    <span className="text-xs font-semibold">{status}</span>
                </div>
            </div>

            {/* CÜZDAN KARTLARI (GRID) */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 w-full max-w-4xl mb-12">
                {/* USDT Kartı */}
                <div className="bg-[#151515] border border-gray-800 rounded-2xl p-6 relative overflow-hidden group hover:border-green-500/50 transition-colors duration-300">
                    <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
                        <span className="text-6xl">💲</span>
                    </div>
                    <p className="text-xs text-gray-500 font-bold tracking-[0.2em] mb-2">USDT BAKİYE</p>
                    <div className="text-3xl md:text-4xl font-bold text-white flex items-baseline gap-1">
                        {wallet.usdt.toFixed(2)} <span className="text-lg text-gray-600 font-medium">USDT</span>
                    </div>
                </div>

                {/* BTC Kartı */}
                <div className="bg-[#151515] border border-gray-800 rounded-2xl p-6 relative overflow-hidden group hover:border-yellow-500/50 transition-colors duration-300">
                    <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:opacity-20 transition-opacity">
                        <span className="text-6xl">₿</span>
                    </div>
                    <p className="text-xs text-gray-500 font-bold tracking-[0.2em] mb-2">KRİPTO VARLIK</p>
                    <div className="text-3xl md:text-4xl font-bold text-yellow-400 flex items-baseline gap-1">
                        {wallet.btc.toFixed(5)} <span className="text-lg text-yellow-700 font-medium">BTC</span>
                    </div>
                </div>
            </div>

            {/* FİYAT GÖSTERGESİ (DEV EKRAN) */}
            <div className="relative group mb-12 cursor-default">
                {/* Arkadaki Glow Efekti */}
                <div className="absolute -inset-1 bg-gradient-to-r from-yellow-600/20 to-orange-600/20 rounded-2xl blur-xl opacity-50 group-hover:opacity-75 transition duration-1000"></div>
                
                <div className="relative bg-black border border-gray-800 rounded-2xl px-16 py-8 flex flex-col items-center shadow-2xl">
                    <span className="text-gray-600 text-xs font-bold tracking-widest mb-2">BTC / USDT FİYATI</span>
                    <div className="text-6xl md:text-8xl font-black text-yellow-400 tabular-nums tracking-tighter drop-shadow-[0_0_15px_rgba(234,179,8,0.3)]">
                        ${price}
                    </div>
                </div>
            </div>

            {/* LOG PANELİ (TERMINAL GÖRÜNÜMÜ) */}
            <div className="w-full max-w-4xl flex-1 min-h-[300px] bg-[#050505] border border-gray-800 rounded-xl overflow-hidden shadow-2xl flex flex-col">
                {/* Terminal Başlığı */}
                <div className="bg-[#1a1a1a] px-4 py-2 border-b border-gray-800 flex justify-between items-center">
                    <div className="flex items-center gap-2">
                        <div className="flex gap-1.5">
                            <div className="w-3 h-3 rounded-full bg-red-500/20 border border-red-500/50"></div>
                            <div className="w-3 h-3 rounded-full bg-yellow-500/20 border border-yellow-500/50"></div>
                            <div className="w-3 h-3 rounded-full bg-green-500/20 border border-green-500/50"></div>
                        </div>
                        <span className="text-xs font-semibold text-gray-400 ml-2">system_logs.log</span>
                    </div>
                    <span className="text-[10px] text-gray-600 uppercase tracking-wider">Live Feed</span>
                </div>
                
                {/* Log İçeriği */}
                <div className="flex-1 p-4 overflow-y-auto font-mono text-xs md:text-sm custom-scrollbar space-y-1.5">
                    {logs.length === 0 && (
                        <div className="text-gray-700 text-center mt-20 animate-pulse">
                            Veri akışı bekleniyor... <br/>
                            (Bağlantı kuruldu, sinyaller dinleniyor)
                        </div>
                    )}
                    
                    {logs.map((log) => (
                        <div key={log.id} className="flex gap-3 hover:bg-white/5 p-1 rounded transition-colors">
                            <span className="text-gray-600 min-w-[80px] select-none">[{log.time}]</span>
                            <span className={`${getLogClass(log.type)} break-all`}>
                                {log.type === 'buy' || log.type === 'sell' ? '⚡ ' : '> '}
                                {log.msg}
                            </span>
                        </div>
                    ))}
                </div>
            </div>

            <div className="mt-4 text-[10px] text-gray-700">
                v2.0.1 Stable • Powered by GoFiber & React
            </div>
        </div>
    );
};

export default TradingMonitor;