import { Navigation } from "@/components/Navigation";
import { Footer } from "@/components/Footer";
import { motion } from "framer-motion";
import { Link } from "wouter";
import { Button } from "@/components/ui/button";
import { useUser } from "@/hooks/use-user";
import { useQuery } from "@tanstack/react-query";
import {
  Users,
  User as UserIcon,
  TrendingUp,
  TrendingDown,
  Wallet,
  Activity,
  Zap,
  ArrowUpRight,
  LogIn
} from "lucide-react";

import { useState } from "react";
import { useToast } from "@/hooks/use-toast";
import { useQueryClient } from "@tanstack/react-query";

interface Bet {
  id: string;
  bet_amount: number;
  potential_win: number;
  status: string;
  result: string | null;
  odds: number;
}

interface CommunityStats {
  highestBalance: { value: number; player: string };
  bestWinRate: { value: number; player: string };
  lowestPnl: { value: number; player: string };
  highestPnl: { value: number; player: string };
  totalPlayers: number;
  totalMatches: number;
  totalBets: number;
}

function useBets() {
  return useQuery({
    queryKey: ["user-bets"],
    queryFn: async () => {
      const accessToken = localStorage.getItem('access_token');
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
      };

      if (accessToken) {
        headers['Authorization'] = `Bearer ${accessToken}`;
      }

      const res = await fetch("/api/bets", {
        headers,
        credentials: "include"
      });
      if (!res.ok) {
        if (res.status === 401) return null;
        throw new Error("Failed to fetch bets");
      }
      const data = await res.json();
      return data.success ? data.bets as Bet[] : null;
    },
    retry: false,
    enabled: !!localStorage.getItem('access_token'),
  });
}

function useCommunityStats() {
  return useQuery({
    queryKey: ["community-stats"],
    queryFn: async () => {
      const accessToken = localStorage.getItem('access_token');
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
      };

      if (accessToken) {
        headers['Authorization'] = `Bearer ${accessToken}`;
      }

      const [playersRes, matchesRes] = await Promise.all([
        fetch("/api/players", { headers, credentials: "include" }),
        fetch("/api/matches", { headers, credentials: "include" })
      ]);
      
      const playersData = await playersRes.json();
      const matchesData = await matchesRes.json();
      
      const players = playersData.success ? playersData.players : [];
      const matches = matchesData.success ? matchesData.matches : [];
      
      let stats: CommunityStats = {
        highestBalance: { value: 0, player: "-" },
        bestWinRate: { value: 0, player: "-" },
        lowestPnl: { value: 0, player: "-" },
        highestPnl: { value: 0, player: "-" },
        totalPlayers: players.length,
        totalMatches: matches.length,
        totalBets: 0
      };
      
      if (players.length > 0) {
        let totalBets = 0;
        let highestBalance = { value: -Infinity, player: "" };
        let bestWinRate = { value: -Infinity, player: "" };
        let lowestPnl = { value: Infinity, player: "" };
        let highestPnl = { value: -Infinity, player: "" };
        
        for (const p of players) {
          totalBets += p.bets || 0;
          const pnl = p.money - (p.topup * 10000);
          const winRate = p.settled_bets > 0 ? (p.won_bets / p.settled_bets) * 100 : 0;
          
          if (p.money > highestBalance.value) {
            highestBalance = { value: p.money, player: p.nickname };
          }
          if (winRate > bestWinRate.value && p.settled_bets > 0) {
            bestWinRate = { value: winRate, player: p.nickname };
          }
          if (pnl < lowestPnl.value) {
            lowestPnl = { value: pnl, player: p.nickname };
          }
          if (pnl > highestPnl.value) {
            highestPnl = { value: pnl, player: p.nickname };
          }
        }
        
        stats = {
          ...stats,
          highestBalance,
          bestWinRate,
          lowestPnl,
          highestPnl,
          totalBets
        };
      }
      
      return stats;
    },
    staleTime: 30000,
  });
}

export default function Dashboard() {
  const { data: user, isLoading: userLoading } = useUser();
  const { data: bets } = useBets();
  const { data: communityStats, isLoading: statsLoading } = useCommunityStats();

  const { toast } = useToast();
  const queryClient = useQueryClient();
  

  const personalStats = {
    totalBets: user?.bets || 0,
    wonBets: user?.won_bets || 0,
    winRate: user?.settled_bets ? ((user.won_bets / user.settled_bets) * 100).toFixed(1) : "0",
    pnl: user ? user.money - (user.topup * 10000) : 0,
    pendingAmount: bets?.filter(b => b.status === 'pending').reduce((sum, b) => sum + b.bet_amount, 0) || 0,
    potentialWin: bets?.filter(b => b.status === 'pending').reduce((sum, b) => sum + b.potential_win, 0) || 0,
    lowestOddsLost: bets?.filter(b => b.status === 'lost').reduce((min, b) => Math.min(min, b.odds), Infinity) || 0,
    highestOddsWon: bets?.filter(b => b.status === 'won').reduce((max, b) => Math.max(max, b.odds), 0) || 0,
  };
  return (
    <div className="min-h-screen flex flex-col bg-background font-sans selection:bg-primary/20">
      <Navigation />
      
      <main className="flex-1 py-12 md:py-20">
        <div className="container mx-auto px-4 max-w-6xl">
          <motion.div 
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="mb-12"
          >
            <h1 className="text-4xl font-display font-bold tracking-tight text-foreground mb-2">
              Dashboard
            </h1>
            <p className="text-muted-foreground max-w-2xl">
              Betting statistics and insights
            </p>
          </motion.div>

          {/* Statistics Sections */}
          <div className="flex flex-col gap-16 mb-20">
            
            {/* Personal Statistics */}
            <section className="space-y-8">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-primary/10">
                  <UserIcon className="size-6 text-primary" />
                </div>
                <h2 className="text-2xl font-black font-display tracking-tight uppercase">Personal Statistics</h2>
              </div>
              
              {userLoading ? (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                  {[1, 2, 3, 4].map(i => (
                    <div key={i} className="p-8 rounded-2xl bg-card/50 animate-pulse h-32" />
                  ))}
                </div>
              ) : user ? (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                  <SimpleStat label="Total Bets" value={personalStats.totalBets.toString()} />
                  <SimpleStat label="Won Bets" value={personalStats.wonBets.toString()} />
                  <SimpleStat label="Win Rate" value={`${personalStats.winRate}%`} />
                  <SimpleStat label="PnL" value={personalStats.pnl.toLocaleString()} />
                  <div className="md:col-span-2 lg:col-span-4 bg-transparent grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 relative overflow-visible">
                    <StatRow label="Pending Amount" value={personalStats.pendingAmount.toLocaleString()} />
                    <StatRow label="Potential Win" value={personalStats.potentialWin.toLocaleString()} variant="outline" />
                    <StatRow label="Lowest Odds Lost" value={personalStats.lowestOddsLost === Infinity ? "-" : personalStats.lowestOddsLost.toFixed(2)} variant="red" />
                    <StatRow label="Highest Odds Won" value={personalStats.highestOddsWon === 0 ? "-" : personalStats.highestOddsWon.toFixed(2)} variant="green" />
                  </div>
                </div>
              ) : (
                <div className="p-12 rounded-2xl border-2 border-dashed border-primary/20 bg-primary/5 text-center">
                  <LogIn className="size-12 text-primary/40 mx-auto mb-4" />
                  <h3 className="text-xl font-bold text-foreground mb-2">Sign in to see your stats</h3>
                  <p className="text-muted-foreground mb-6">Create an account or log in to track your betting performance</p>
                  <div className="flex justify-center">
                    <Button
                      size="lg"
                      className="rounded-full px-8 bg-gradient-to-r from-primary to-primary/80 hover:from-primary/90 hover:to-primary shadow-lg hover:shadow-xl transition-all duration-300 transform hover:scale-[1.02] flex items-center justify-center font-bold uppercase tracking-widest"
                      onClick={() => window.dispatchEvent(new CustomEvent('openAuthModal'))}
                    >
                      Get Started
                    </Button>
                  </div>
                </div>
              )}
            </section>

            {/* Community Statistics */}
            <section className="space-y-8">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-lg bg-primary/10">
                  <Users className="size-6 text-primary" />
                </div>
                <h2 className="text-2xl font-black font-display tracking-tight uppercase">Community Statistics</h2>
              </div>
              
              {statsLoading ? (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                  {[1, 2, 3, 4].map(i => (
                    <div key={i} className="p-6 rounded-2xl bg-card/50 animate-pulse h-32" />
                  ))}
                </div>
              ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                  <StatCard 
                    label="Highest Balance" 
                    value={`${(communityStats?.highestBalance.value ?? 0).toLocaleString()}`} 
                    subtext={communityStats?.highestBalance.player || "-"} 
                    icon={<Wallet className="size-32" />} 
                  />
                  <StatCard 
                    label="Best Win Rate" 
                    value={`${(communityStats?.bestWinRate.value && communityStats.bestWinRate.value !== -Infinity ? communityStats.bestWinRate.value : 0).toFixed(1)}%`} 
                    subtext={communityStats?.bestWinRate.player || "-"} 
                    icon={<Zap className="size-32" />} 
                  />
                  <StatCard 
                    label="Lowest PnL" 
                    value={(communityStats?.lowestPnl.value && communityStats.lowestPnl.value !== Infinity ? communityStats.lowestPnl.value : 0).toLocaleString()} 
                    subtext={communityStats?.lowestPnl.player || "-"} 
                    icon={<TrendingDown className="size-32" />} 
                  />
                  <StatCard 
                    label="Highest PnL" 
                    value={(communityStats?.highestPnl.value && communityStats.highestPnl.value !== -Infinity ? communityStats.highestPnl.value : 0).toLocaleString()} 
                    subtext={communityStats?.highestPnl.player || "-"} 
                    icon={<TrendingUp className="size-32" />} 
                  />
                </div>
              )}
            </section>

            {/* Platform Overview */}
            <motion.section 
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: 0.4 }}
              className="space-y-8"
            >
              <div className="flex items-center gap-3 px-2">
                <div className="p-2 rounded-lg bg-primary/10">
                  <Activity className="size-6 text-primary" />
                </div>
                <h2 className="text-2xl font-black font-display tracking-tight uppercase">Platform Overview</h2>
              </div>
              
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <Link href="/leaderboard">
                  <motion.div 
                    whileHover={{ y: -5, scale: 1.02 }}
                    className="flex flex-col items-center justify-center text-center p-12 bg-primary/5 rounded-3xl border border-primary/10 shadow-xl relative overflow-hidden cursor-pointer"
                  >
                    <div className="absolute top-4 right-4">
                      <ArrowUpRight className="size-5 text-primary/40 group-hover:text-primary transition-colors" />
                    </div>
                    <span className="text-8xl font-light text-primary tracking-tighter mb-2">{communityStats?.totalPlayers || 0}</span>
                    <span className="text-xs font-bold uppercase tracking-[0.3em] text-primary/60">Active Players</span>
                  </motion.div>
                </Link>
                
                <Link href="/">
                  <motion.div 
                    whileHover={{ y: -5, scale: 1.02 }}
                    className="flex flex-col items-center justify-center text-center p-12 bg-primary/10 rounded-3xl border border-primary/20 shadow-xl relative overflow-hidden cursor-pointer"
                  >
                    <div className="absolute top-4 right-4">
                      <ArrowUpRight className="size-5 text-primary/40 group-hover:text-primary transition-colors" />
                    </div>
                    <span className="text-8xl font-light text-primary tracking-tighter mb-2">{communityStats?.totalMatches || 0}</span>
                    <span className="text-xs font-bold uppercase tracking-[0.3em] text-primary/70">Available Matches</span>
                  </motion.div>
                </Link>
                
                <motion.div 
                  whileHover={{ y: -5, scale: 1.02 }}
                  className="flex flex-col items-center justify-center text-center p-12 bg-primary/20 rounded-3xl border border-primary/30 shadow-xl"
                >
                  <span className="text-8xl font-light text-primary tracking-tighter mb-2">{communityStats?.totalBets || 0}</span>
                  <span className="text-xs font-bold uppercase tracking-[0.3em] text-primary/80">Total Bets</span>
                </motion.div>
              </div>
            </motion.section>
          </div>
        </div>
      </main>

      <Footer />

    </div>
  );
}

function StatCard({ label, value, subtext, icon }: { label: string, value: string, subtext: string, icon: React.ReactNode }) {
  return (
    <motion.div 
      whileHover={{ y: -5, scale: 1.02 }}
      className="p-6 rounded-2xl border border-border/50 bg-card/50 backdrop-blur-sm flex items-center justify-between shadow-lg overflow-hidden relative group"
    >
      <div className="flex flex-col relative z-10">
        <span className="text-[10px] uppercase font-bold tracking-[0.2em] text-muted-foreground mb-1">{label}</span>
        <span className="text-3xl font-extrabold text-foreground tracking-tight">{value}</span>
        <div className="flex items-center gap-1.5 text-muted-foreground/80 mt-1">
          <UserIcon className="size-3" />
          <span className="text-[11px] font-medium">{subtext}</span>
        </div>
      </div>
      <div className="absolute -right-2 -bottom-8 text-primary/5 group-hover:text-primary/10 transition-all duration-500 rotate-12 pointer-events-none">
        {icon}
      </div>
    </motion.div>
  );
}

function SimpleStat({ label, value }: { label: string, value: string }) {
  return (
    <motion.div 
      whileHover={{ y: -5, scale: 1.02 }}
      className="p-8 rounded-2xl bg-card/50 backdrop-blur-sm border border-border/50 flex flex-col items-center justify-center text-center shadow-lg group"
    >
      <span className="text-[10px] uppercase font-bold text-muted-foreground tracking-[0.2em] mb-3">{label}</span>
      <span className="text-5xl font-light tracking-tighter text-foreground">{value}</span>
    </motion.div>
  );
}

function StatRow({ label, value, variant = "primary" }: { label: string, value: string, variant?: "primary" | "red" | "green" | "outline" }) {
  const variants = {
    primary: "bg-primary text-primary-foreground border-primary/30",
    red: "bg-red-600 text-white border-red-400/30",
    green: "bg-green-600 text-white border-green-400/30",
    outline: "bg-transparent border-primary text-primary"
  };

  return (
    <motion.div 
      whileHover={{ scale: 1.05 }}
      className={`flex flex-col items-center justify-center p-6 ${variants[variant]} rounded-2xl shadow-xl group transition-all duration-300 border-4`}
    >
      <span className="text-3xl font-light tracking-tighter mb-2">{value}</span>
      <span className="text-[10px] font-bold uppercase tracking-[0.2em] opacity-80">{label}</span>
    </motion.div>
  );
}

function StatLine({ label, value }: { label: string, value: string }) {
  return (
    <div className="flex items-center justify-between py-3 border-b border-border/30 last:border-0 md:border-b-0 group">
      <span className="text-xs font-bold text-muted-foreground uppercase tracking-widest group-hover:text-primary/70 transition-colors">{label}</span>
      <span className="text-xl font-extrabold text-primary tracking-tight group-hover:scale-110 transition-transform">{value}</span>
    </div>
  );
}
