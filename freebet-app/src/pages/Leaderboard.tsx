import { Navigation } from "@/components/Navigation";
import { Footer } from "@/components/Footer";
import { PlayerCard } from "@/components/PlayerCard";
import { HelmetManager } from "@/components/HelmetManager";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { motion } from "framer-motion";
import { useState, useMemo, useEffect } from "react";
import { ArrowUpDown, ChevronLeft, ChevronRight, Loader2, Users, Wallet, TrendingUp, TrendingDown, Zap, User as UserIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";


interface Player {
  id: string;
  nickname: string;
  money: number;
  bets: number;
  won_bets: number;
  settled_bets: number;
  avg_odds: number;
  topup: number;
  created: string;
  updated: string;
}

interface PlayersResponse {
  success: boolean;
  players: Player[];
  pagination: {
    limit: number;
    offset: number;
    total: number;
    has_more: boolean;
  };
}

type SortKey = "rank" | "balance" | "bets" | "winRate" | "avgOdds" | "topUps" | "pnl" | "registeredAt";

interface CommunityStats {
  highestBalance: { value: number; player: string };
  bestWinRate: { value: number; player: string };
  lowestPnl: { value: number; player: string };
  highestPnl: { value: number; player: string };
  totalPlayers: number;
  totalMatches: number;
  totalBets: number;
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

export default function Leaderboard() {
  const [players, setPlayers] = useState<Player[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [sortConfig, setSortConfig] = useState<{ key: SortKey; direction: 'asc' | 'desc' } | null>({ key: 'pnl', direction: 'desc' });
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 10;

  const { data: communityStats, isLoading: statsLoading } = useCommunityStats();

  useEffect(() => {
    const fetchPlayers = async () => {
      try {
        setLoading(true);
        const response = await fetch('/api/players');
        const data: PlayersResponse = await response.json();
        if (data.success) {
          setPlayers(data.players || []);
        } else {
          setError('Failed to load leaderboard data');
        }
      } catch (err) {
        console.error('Error fetching players:', err);
        setError('Failed to connect to server');
      } finally {
        setLoading(false);
      }
    };
    fetchPlayers();
  }, []);

  const transformedData = useMemo(() => {
    return players.map((player) => {
      const winRate = player.settled_bets > 0
        ? Math.round((player.won_bets / player.settled_bets) * 100)
        : 0;
      const pnl = Math.round(player.money - (player.topup * 10000));
      return {
        id: player.id,
        name: player.nickname,
        balance: player.money,
        bets: player.bets,
        winRate,
        avgOdds: Math.round(player.avg_odds * 100) / 100 || 0,
        topUps: player.topup,
        pnl,
        registeredAt: player.created
      };
    });
  }, [players]);

  const sortedData = useMemo(() => {
    let items = [...transformedData];
    if (sortConfig !== null) {
      items.sort((a: any, b: any) => {
        if (a[sortConfig.key] < b[sortConfig.key]) {
          return sortConfig.direction === 'asc' ? -1 : 1;
        }
        if (a[sortConfig.key] > b[sortConfig.key]) {
          return sortConfig.direction === 'asc' ? 1 : -1;
        }
        return 0;
      });
    }
    // Assign ranks based on sorted order
    return items.map((item, index) => ({
      ...item,
      rank: index + 1
    }));
  }, [transformedData, sortConfig]);

  const totalPages = Math.ceil(sortedData.length / itemsPerPage);
  const paginatedData = useMemo(() => {
    return sortedData.slice((currentPage - 1) * itemsPerPage, currentPage * itemsPerPage);
  }, [sortedData, currentPage]);

  const requestSort = (key: SortKey) => {
    let direction: 'asc' | 'desc' = 'desc';
    if (sortConfig && sortConfig.key === key && sortConfig.direction === 'desc') {
      direction = 'asc';
    }
    setSortConfig({ key, direction });
  };


  if (loading) {
    return (
      <div className="min-h-screen flex flex-col bg-background font-sans selection:bg-primary/20">
        <Navigation />
        <main className="flex-1 py-12 md:py-20 flex items-center justify-center">
          <Loader2 className="size-8 animate-spin text-primary" />
        </main>
        <Footer />
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex flex-col bg-background font-sans selection:bg-primary/20">
        <Navigation />
        <main className="flex-1 py-12 md:py-20 flex items-center justify-center">
          <p className="text-red-500">{error}</p>
        </main>
        <Footer />
      </div>
    );
  }

  return (
    <>
      <HelmetManager
        title="Leaderboard - Top Virtual Bettors | FreeBet Guru"
        description="Check out the top virtual bettors on FreeBet Guru. See player rankings, win rates, profit/loss statistics, and betting performance metrics."
        keywords="betting leaderboard, top bettors, player rankings, betting statistics, profit loss, virtual betting rankings, betting performance"
      />
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
              Leaderboard
            </h1>
            <p className="text-muted-foreground max-w-2xl">
              Top performers in our betting community.
            </p>
          </motion.div>

          {/* Community Statistics */}
          <section className="space-y-8 mb-16">

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


          {paginatedData.length === 0 ? (
            <div className="bg-muted/50 border border-border/50 rounded-2xl p-12 text-center">
              <div className="text-muted-foreground text-lg mb-2">🏆 No players yet</div>
              <div className="text-muted-foreground/80">Be the first to join the leaderboard!</div>
            </div>
          ) : (
            <>
              {/* Mobile Cards View */}
              <div className="lg:hidden grid grid-cols-1 gap-6">
                {paginatedData.map((player) => (
                  <PlayerCard key={player.id} player={player} />
                ))}
              </div>

              {/* Desktop Table View */}
              <div className="hidden lg:block bg-card border border-border/50 rounded-2xl overflow-hidden shadow-sm mb-6">
                <Table>
                  <TableHeader className="bg-muted/50">
                    <TableRow className="border-border/50">
                      <TableHead className="w-[80px] font-medium text-xs uppercase tracking-wider text-center py-4">
                        <Button variant="ghost" onClick={() => requestSort('rank')} className="hover:bg-transparent p-0 font-medium text-xs uppercase tracking-wider gap-2 mx-auto">
                          Rank
                          <ArrowUpDown className="size-3" />
                        </Button>
                      </TableHead>
                      <TableHead className="font-medium text-xs uppercase tracking-wider py-4">Player</TableHead>
                      <TableHead className="font-medium text-xs uppercase tracking-wider py-4">
                        <Button variant="ghost" onClick={() => requestSort('registeredAt')} className="hover:bg-transparent p-0 font-medium text-xs uppercase tracking-wider gap-2">
                          Joined
                          <ArrowUpDown className="size-3" />
                        </Button>
                      </TableHead>
                      <TableHead className="font-medium text-xs uppercase tracking-wider py-4 text-right">
                        <Button variant="ghost" onClick={() => requestSort('balance')} className="hover:bg-transparent p-0 font-medium text-xs uppercase tracking-wider gap-2 ml-auto">
                          Balance
                          <ArrowUpDown className="size-3" />
                        </Button>
                      </TableHead>
                      <TableHead className="font-medium text-xs uppercase tracking-wider py-4 text-center">
                        <Button variant="ghost" onClick={() => requestSort('bets')} className="hover:bg-transparent p-0 font-medium text-xs uppercase tracking-wider gap-2 mx-auto">
                          Bets
                          <ArrowUpDown className="size-3" />
                        </Button>
                      </TableHead>
                      <TableHead className="font-medium text-xs uppercase tracking-wider py-4 text-center">
                        <Button variant="ghost" onClick={() => requestSort('winRate')} className="hover:bg-transparent p-0 font-medium text-xs uppercase tracking-wider gap-2 mx-auto">
                          Win Rate
                          <ArrowUpDown className="size-3" />
                        </Button>
                      </TableHead>
                      <TableHead className="font-medium text-xs uppercase tracking-wider py-4 text-center">
                        <Button variant="ghost" onClick={() => requestSort('avgOdds')} className="hover:bg-transparent p-0 font-medium text-xs uppercase tracking-wider gap-2 mx-auto">
                          Avg Odds
                          <ArrowUpDown className="size-3" />
                        </Button>
                      </TableHead>
                      <TableHead className="font-medium text-xs uppercase tracking-wider py-4 text-center">
                        <Button variant="ghost" onClick={() => requestSort('topUps')} className="hover:bg-transparent p-0 font-medium text-xs uppercase tracking-wider gap-2 mx-auto">
                          Top-ups
                          <ArrowUpDown className="size-3" />
                        </Button>
                      </TableHead>
                      <TableHead className="font-medium text-xs uppercase tracking-wider py-4 text-right">
                        <Button variant="ghost" onClick={() => requestSort('pnl')} className="hover:bg-transparent p-0 font-medium text-xs uppercase tracking-wider gap-2 ml-auto">
                          PnL
                          <ArrowUpDown className="size-3" />
                        </Button>
                      </TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedData.map((player) => (
                      <TableRow 
                        key={player.id} 
                        className="border-border/50 hover:bg-muted/30 transition-colors cursor-pointer"
                        onClick={() => window.location.href = `/player/${encodeURIComponent(player.name)}`}
                      >
                        <TableCell className="text-center font-black text-muted-foreground text-sm py-4">
                          {player.rank === 1 ? "🥇" : player.rank === 2 ? "🥈" : player.rank === 3 ? "🥉" : player.rank}
                        </TableCell>
                        <TableCell className="py-4">
                          <Link href={`/player/${encodeURIComponent(player.name)}`}>
                            <span className="font-semibold text-sm text-foreground hover:text-primary transition-colors cursor-pointer">{player.name}</span>
                          </Link>
                        </TableCell>
                        <TableCell className="text-muted-foreground/60 font-medium text-[10px] uppercase tracking-wider py-4">
                          {new Date(player.registeredAt).toLocaleDateString('en-US', { month: 'short', year: 'numeric' })}
                        </TableCell>
                        <TableCell className="text-right font-bold text-foreground text-base py-4">
                          {player.balance.toLocaleString()}
                        </TableCell>
                        <TableCell className="text-center font-medium text-muted-foreground text-sm py-4">{player.bets}</TableCell>
                        <TableCell className="text-center py-4">
                          <Badge className="font-bold bg-primary text-primary-foreground text-xs px-3 py-1 rounded-full">
                            {player.winRate}%
                          </Badge>
                        </TableCell>
                        <TableCell className="text-center font-bold text-muted-foreground text-sm py-4">{player.avgOdds}</TableCell>
                        <TableCell className="text-center font-medium text-muted-foreground text-sm py-4">{player.topUps}</TableCell>
                        <TableCell className="text-right py-4">
                          <span className={`font-bold text-base ${player.pnl >= 0 ? 'text-green-600 dark:text-green-500' : 'text-red-600 dark:text-red-500'}`}>
                            {player.pnl >= 0 ? `+${player.pnl.toLocaleString()}` : player.pnl.toLocaleString()}
                          </span>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>

              {/* Pagination */}
              {totalPages > 1 && (
                <div className="flex items-center justify-center gap-1 mt-8 lg:mt-12 py-6 border-t border-border/30">
                  <Button 
                    variant="ghost" 
                    size="icon" 
                    onClick={() => {
                      setCurrentPage(prev => Math.max(1, prev - 1));
                      window.scrollTo({ top: 0, behavior: 'smooth' });
                    }}
                    disabled={currentPage === 1}
                    className="rounded-md hover:bg-transparent transition-all group"
                  >
                    <ChevronLeft className="size-4 group-hover:text-primary group-hover:-translate-x-0.5 transition-transform" />
                  </Button>
                  <div className="flex items-center gap-2 px-4">
                    {Array.from({ length: totalPages }, (_, i) => i + 1).map(page => (
                      <Button
                        key={page}
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          setCurrentPage(page);
                          window.scrollTo({ top: 0, behavior: 'smooth' });
                        }}
                        className={cn(
                          "w-10 h-10 p-0 rounded-md transition-all font-bold text-sm hover:bg-transparent",
                          currentPage === page 
                            ? "text-primary font-bold" 
                            : "text-muted-foreground hover:text-primary"
                        )}
                      >
                        {page}
                      </Button>
                    ))}
                  </div>
                  <Button 
                    variant="ghost" 
                    size="icon" 
                    onClick={() => {
                      setCurrentPage(prev => Math.min(totalPages, prev + 1));
                      window.scrollTo({ top: 0, behavior: 'smooth' });
                    }}
                    disabled={currentPage === totalPages}
                    className="rounded-md hover:bg-transparent transition-all group"
                  >
                    <ChevronRight className="size-4 group-hover:text-primary group-hover:translate-x-0.5 transition-transform" />
                  </Button>
                </div>
              )}
            </>
          )}
        </div>
      </main>

      <Footer />
    </div>
    </>
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
