"use client";
import { useEffect, useState } from "react";

interface PingResult {
  subdomain: string;
  is_up: boolean;
  status_code: number;
  response_time: string;
}

export default function Home() {
  const [results, setResults] = useState<PingResult[]>([]);
  const [loading, setLoading] = useState(true);
  const [lastUpdated, setLastUpdated] = useState<string>("");

  const fetchStatus = () => {
    setLoading(true);
    // Ubah ke "/api/health" jika Next.js rewrites sudah aktif di production
    fetch("https://go-monitoring-api.ryaze.my.id/api/health")
      .then((res) => res.json())
      .then((data) => {
        setResults(data.data || []);
        setLastUpdated(new Date().toLocaleTimeString("id-ID"));
        setLoading(false);
      })
      .catch((err) => {
        console.error("Gagal terhubung ke backend", err);
        setResults([]);
        setLoading(false);
      });
  };

  useEffect(() => {
    fetchStatus();
    const interval = setInterval(fetchStatus, 15000);
    return () => clearInterval(interval);
  }, []);

  return (
    <main className="min-h-screen bg-[#0B0F19] text-gray-100 font-sans selection:bg-blue-500/30">
      {/* Background Glow Effect */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-full max-w-lg h-[400px] bg-blue-600/20 blur-[120px] pointer-events-none rounded-full" />

      <div className="max-w-6xl mx-auto p-6 md:p-10 relative z-10">
        {/* Header Section */}
        <header className="flex flex-col md:flex-row justify-between items-start md:items-center mb-12 gap-4">
          <div>
            <h1 className="text-3xl md:text-4xl font-extrabold tracking-tight bg-gradient-to-r from-blue-400 to-emerald-400 bg-clip-text text-transparent">
              Monitoring Subdomain
            </h1>
            <p className="text-gray-400 mt-2 text-sm md:text-base flex items-center gap-2">
              <svg
                className="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              Pembaruan terakhir: {lastUpdated || "Menunggu..."}
            </p>
          </div>

          <button
            onClick={fetchStatus}
            disabled={loading}
            className="flex items-center gap-2 bg-white/5 hover:bg-white/10 border border-white/10 px-5 py-2.5 rounded-xl font-medium transition-all active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed shadow-sm backdrop-blur-sm"
          >
            <svg
              className={`w-4 h-4 ${loading ? "animate-spin text-blue-400" : "text-gray-300"}`}
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
              />
            </svg>
            {loading ? "Menyegarkan..." : "Refresh Status"}
          </button>
        </header>

        {/* Loading Skeleton */}
        {loading && results?.length === 0 ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
            {[1, 2, 3, 4, 5, 6].map((n) => (
              <div
                key={n}
                className="bg-white/5 border border-white/5 p-6 rounded-2xl animate-pulse"
              >
                <div className="flex justify-between items-start mb-4">
                  <div className="h-5 bg-white/10 rounded w-1/2"></div>
                  <div className="h-5 bg-white/10 rounded-full w-16"></div>
                </div>
                <div className="h-4 bg-white/10 rounded w-1/3 mb-2"></div>
                <div className="h-4 bg-white/10 rounded w-1/4"></div>
              </div>
            ))}
          </div>
        ) : (
          /* Cards Grid */
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
            {results?.map((item, index) => {
              // Membersihkan "https://" agar tampilan nama subdomain lebih rapi
              const cleanName = item.subdomain.replace(/^https?:\/\//, "");

              return (
                <div
                  key={index}
                  className="group relative bg-[#131825] border border-gray-800 hover:border-gray-600 p-6 rounded-2xl flex flex-col gap-4 shadow-lg hover:shadow-xl transition-all duration-300 hover:-translate-y-1"
                >
                  <div className="flex justify-between items-start">
                    <a
                      href={item.subdomain}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="font-semibold text-lg text-gray-100 break-all mr-3 group-hover:text-blue-400 transition-colors line-clamp-2"
                      title={item.subdomain}
                    >
                      {cleanName}
                    </a>

                    {item.is_up ? (
                      <div className="flex items-center gap-2 bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 px-3 py-1.5 rounded-full text-xs font-bold whitespace-nowrap shadow-[0_0_10px_rgba(16,185,129,0.1)]">
                        <span className="relative flex h-2 w-2">
                          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                          <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
                        </span>
                        UP
                      </div>
                    ) : (
                      <div className="flex items-center gap-2 bg-red-500/10 border border-red-500/20 text-red-400 px-3 py-1.5 rounded-full text-xs font-bold whitespace-nowrap shadow-[0_0_10px_rgba(239,68,68,0.1)]">
                        <span className="relative flex h-2 w-2">
                          <span className="relative inline-flex rounded-full h-2 w-2 bg-red-500"></span>
                        </span>
                        DOWN
                      </div>
                    )}
                  </div>

                  <div className="grid grid-cols-2 gap-3 mt-auto pt-4 border-t border-gray-800/50">
                    <div>
                      <p className="text-xs text-gray-500 mb-1 uppercase tracking-wider font-semibold">
                        Status Code
                      </p>
                      <p className="text-sm font-mono text-gray-300 flex items-center gap-1.5">
                        <svg
                          className="w-3.5 h-3.5 text-gray-500"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth="2"
                            d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                          />
                        </svg>
                        {item.status_code || "N/A"}
                      </p>
                    </div>
                    <div>
                      <p className="text-xs text-gray-500 mb-1 uppercase tracking-wider font-semibold">
                        Response Time
                      </p>
                      <p className="text-sm font-mono text-gray-300 flex items-center gap-1.5">
                        <svg
                          className="w-3.5 h-3.5 text-gray-500"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth="2"
                            d="M13 10V3L4 14h7v7l9-11h-7z"
                          />
                        </svg>
                        {item.response_time}
                      </p>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}

        {/* Empty State */}
        {!loading && results?.length === 0 && (
          <div className="text-center py-20 bg-white/5 border border-white/5 rounded-2xl border-dashed">
            <svg
              className="w-12 h-12 text-gray-600 mx-auto mb-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
            <h3 className="text-lg font-medium text-gray-300">
              Belum ada subdomain
            </h3>
            <p className="text-gray-500 mt-1">
              Sistem belum menemukan data monitoring dari server backend.
            </p>
          </div>
        )}
      </div>
    </main>
  );
}
