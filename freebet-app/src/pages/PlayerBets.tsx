import { Navigation } from "@/components/Navigation";
import { Footer } from "@/components/Footer";
import { useQuery } from "@tanstack/react-query";
import { Bet } from "@/types";
import { format } from "date-fns";
import { Badge } from "@/components/ui/badge";
import { motion } from "framer-motion";
import { useState, useMemo } from "react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { ChevronLeft, ChevronRight, ArrowUpDown, ArrowLeft } from "lucide-react";
import { BetCard } from "@/components/BetCard";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useParams, Link } from "wouter";
import jsPDF from 'jspdf';
import { Download, FileText } from "lucide-react";

interface PlayerBetsResponse {
  success: boolean;
  player: {
    id: string;
    nickname: string;
    money: number;
    created: string;
  };
  bets: Bet[];
  stats: {
    total_bets: number;
    won_bets: number;
    settled_bets: number;
    win_rate: number;
    avg_odds: number;
  };
}

type SortConfig = {
  key: keyof Bet | 'potentialWin';
  direction: 'asc' | 'desc';
} | null;

export default function PlayerBets() {
  const params = useParams<{ nickname: string }>();
  const playerNickname = params.nickname;

  const { data, isLoading, error } = useQuery<PlayerBetsResponse>({
    queryKey: [`/api/bets?player=${encodeURIComponent(playerNickname || '')}`],
    queryFn: async () => {
      const res = await fetch(`/api/bets?player=${encodeURIComponent(playerNickname || '')}`);
      if (!res.ok) throw new Error("Failed to fetch player bets");
      return res.json();
    },
    enabled: !!playerNickname,
  });

  const [sortConfig, setSortConfig] = useState<SortConfig>(null);
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [currentPage, setCurrentPage] = useState(1);
  const itemsPerPage = 20;

  const processedBets = useMemo(() => {
    if (!data?.bets || !Array.isArray(data.bets)) return [];
    
    let filtered = [...data.bets];
    if (statusFilter !== "all") {
      filtered = filtered.filter(b => b.status === statusFilter);
    }

    if (sortConfig) {
      filtered.sort((a, b) => {
        let aValue: any;
        let bValue: any;

        if (sortConfig.key === 'potentialWin') {
          aValue = Number(a.potential_win || 0);
          bValue = Number(b.potential_win || 0);
        } else if (sortConfig.key === 'bet_amount' || sortConfig.key === 'odds') {
          aValue = Number(a[sortConfig.key] || 0);
          bValue = Number(b[sortConfig.key] || 0);
        } else {
          aValue = a[sortConfig.key as keyof Bet];
          bValue = b[sortConfig.key as keyof Bet];
        }

        if (aValue < bValue) return sortConfig.direction === 'asc' ? -1 : 1;
        if (aValue > bValue) return sortConfig.direction === 'asc' ? 1 : -1;
        return 0;
      });
    }

    return filtered;
  }, [data?.bets, sortConfig, statusFilter]);

  const totalPages = useMemo(() => {
    if (!Array.isArray(processedBets) || processedBets.length === 0) return 0;
    return Math.ceil(processedBets.length / itemsPerPage);
  }, [processedBets, itemsPerPage]);

  const paginatedBets = useMemo(() => {
    if (!Array.isArray(processedBets)) return [];
    return processedBets.slice(
      (currentPage - 1) * itemsPerPage,
      currentPage * itemsPerPage
    );
  }, [processedBets, currentPage, itemsPerPage]);

  const lowestOddsLost = useMemo(() => {
    if (!data?.bets || !Array.isArray(data.bets)) return null;
    const lostBets = data.bets.filter(b => b.status === 'lost');
    if (lostBets.length === 0) return null;
    return Math.min(...lostBets.map(b => Number(b.odds || 0)));
  }, [data?.bets]);

  const highestOddsWon = useMemo(() => {
    if (!data?.bets || !Array.isArray(data.bets)) return null;
    const wonBets = data.bets.filter(b => b.status === 'won');
    if (wonBets.length === 0) return null;
    return Math.max(...wonBets.map(b => Number(b.odds || 0)));
  }, [data?.bets]);

  const handleSort = (key: any) => {
    setSortConfig(prev => {
      if (!prev) return { key, direction: 'asc' };
      const direction = prev.key === key && prev.direction === 'asc' ? 'desc' : 'asc';
      return { key, direction };
    });
  };

  const generatePDF = async () => {
    if (!paginatedBets || paginatedBets.length === 0) {
      return;
    }

    const pdf = new jsPDF('landscape');
    const pageWidth = pdf.internal.pageSize.getWidth();
    const pageHeight = pdf.internal.pageSize.getHeight();
    const margin = 10;
    let currentY = margin;

    // Table headers
    pdf.setFontSize(10);
    pdf.setFont('helvetica', 'bold');
    const headers = ['#', 'Bet Date', 'Match', 'Pick', 'Stake', 'Odds', 'Potential', 'Status'];
    const columnWidths = [10, 35, 160, 20, 20, 15, 30, 20];

    // Left-aligned columns (#, Bet Date, Match)
    for (let i = 0; i < 3; i++) {
      let xPos = margin;
      for (let j = 0; j < i; j++) {
        xPos += columnWidths[j];
      }
      pdf.text(headers[i], xPos, currentY);
    }

    // Right-aligned columns (Pick, Stake, Odds, Potential, Status)
    let rightXPos = pageWidth - margin;
    for (let i = headers.length - 1; i >= 3; i--) {
      rightXPos -= columnWidths[i];
      pdf.text(headers[i], rightXPos, currentY);
      if (i > 3) rightXPos -= 5; // Small gap between columns
    }

    currentY += 8;

    // Table data
    pdf.setFont('helvetica', 'normal');
    paginatedBets.forEach((bet, index) => {
      if (currentY > pageHeight - 15) {
        pdf.addPage();
        currentY = margin;
      }

      const rowData = [
        (currentPage - 1) * itemsPerPage + index + 1,
        format(new Date(bet.created_at), "MMM d, yyyy HH:mm"),
        `${bet.home_team} vs ${bet.away_team}`,
        bet.bet_type,
        `${Number(bet.bet_amount || 0).toLocaleString()}`,
        Number(bet.odds || 0).toFixed(2),
        `${Number(bet.potential_win || 0).toLocaleString(undefined, { maximumFractionDigits: 2 })}`,
        bet.status
      ];

      // Left-aligned data (#, Bet Date, Match)
      for (let i = 0; i < 3; i++) {
        let xPos = margin;
        for (let j = 0; j < i; j++) {
          xPos += columnWidths[j];
        }
        pdf.text(String(rowData[i]), xPos, currentY);
      }

      // Right-aligned data (Pick, Stake, Odds, Potential, Status)
      let rightXPos = pageWidth - margin;
      for (let i = rowData.length - 1; i >= 3; i--) {
        rightXPos -= columnWidths[i];
        pdf.text(String(rowData[i]), rightXPos, currentY);
        if (i > 3) rightXPos -= 5; // Small gap between columns
      }

      currentY += 6;
    });


    // Download
    pdf.save(`${player.nickname}_bets_${new Date().toISOString().split('T')[0]}.pdf`);
  };

  const generateCSV = async () => {
    if (!paginatedBets || paginatedBets.length === 0) {
      return;
    }

    // CSV headers
    const headers = ['#', 'Bet Date', 'Match', 'Pick', 'Stake', 'Odds', 'Potential', 'Status'];
    let csvContent = headers.join(',') + '\n';

    // CSV data
    paginatedBets.forEach((bet, index) => {
      const rowData = [
        (currentPage - 1) * itemsPerPage + index + 1,
        format(new Date(bet.created_at), "MMM d, yyyy HH:mm"),
        `"${bet.home_team} vs ${bet.away_team}"`, // Wrap in quotes to handle commas
        bet.bet_type,
        `${Number(bet.bet_amount || 0).toLocaleString()}`,
        Number(bet.odds || 0).toFixed(2),
        `${Number(bet.potential_win || 0).toLocaleString(undefined, { maximumFractionDigits: 2 })}`,
        bet.status
      ];

      csvContent += rowData.join(',') + '\n';
    });

    // Create and download CSV file
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    link.setAttribute('href', url);
    link.setAttribute('download', `${player.nickname}_bets_${new Date().toISOString().split('T')[0]}.csv`);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "won":
        return <Badge className="bg-green-500/10 text-green-600 dark:text-green-400 border-green-500/20 font-semibold text-xs">Won</Badge>;
      case "lost":
        return <Badge className="bg-red-500/10 text-red-600 dark:text-red-400 border-red-500/20 font-semibold text-xs">Lost</Badge>;
      default:
        return <Badge className="bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20 font-semibold text-xs">Pending</Badge>;
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex flex-col bg-background font-sans">
        <Navigation />
        <main className="flex-1 py-12 md:py-20 flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </main>
        <Footer />
      </div>
    );
  }

  if (error || !data?.success) {
    return (
      <div className="min-h-screen flex flex-col bg-background font-sans">
        <Navigation />
        <main className="flex-1 py-12 md:py-20">
          <div className="container mx-auto px-4 max-w-6xl text-center">
            <h1 className="text-2xl font-bold text-foreground mb-4">Player not found</h1>
            <Link href="/leaderboard">
              <Button variant="outline">
                <ArrowLeft className="mr-2 size-4" />
                Back to Leaderboard
              </Button>
            </Link>
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  const { player, stats } = data;

  return (
    <div className="min-h-screen flex flex-col bg-background font-sans selection:bg-primary/20">
      <Navigation />
      
      <main className="flex-1 py-12 md:py-20">
        <div className="container mx-auto px-4 max-w-6xl">
          <motion.div 
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
            className="mb-8"
          >
            <Link href="/leaderboard">
              <Button variant="ghost" size="sm" className="mb-4 -ml-2">
                <ArrowLeft className="mr-2 size-4" />
                Back to Leaderboard
              </Button>
            </Link>

            <div className="flex items-center justify-between mb-2">
              <h1 className="text-3xl lg:text-4xl font-display font-bold tracking-tight text-foreground">
                {player.nickname}'s Bets
              </h1>
              {/* Info icon for mobile to show stats cards */}
              <div className="md:hidden ml-2">
                <button 
                  onClick={() => document.getElementById('mobile-stats-cards')?.classList.toggle('hidden')}
                  className="text-muted-foreground hover:text-foreground focus:outline-none"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <circle cx="12" cy="12" r="10"></circle>
                    <line x1="12" y1="16" x2="12" y2="12"></line>
                    <line x1="12" y1="8" x2="12.01" y2="8"></line>
                  </svg>
                </button>
              </div>
            </div>
            <p className="text-muted-foreground mb-8">
              Member since {format(new Date(player.created), "MMMM yyyy")}
            </p>

            {/* Desktop stats display - hidden on mobile */}
            <div className="hidden md:flex flex-wrap items-center gap-x-6 gap-y-2 text-sm text-muted-foreground mb-8">
              <span><strong className="text-foreground">{stats.total_bets}</strong> bets</span>
              <span className="text-border">|</span>
              <span><strong className="text-green-500">{stats.win_rate.toFixed(1)}%</strong> win rate</span>
              <span className="text-border">|</span>
              <span><strong className="text-foreground">{stats.avg_odds.toFixed(2)}</strong> Avg odds</span>
              <span className="text-border">|</span>
              <span><strong className="text-blue-500">{Number(player.money).toLocaleString()}</strong> Balance</span>
              {lowestOddsLost !== null && (
                <>
                  <span className="text-border">|</span>
                  <span className="flex items-center gap-1">
                    <strong className="text-red-500">{lowestOddsLost.toFixed(2)}</strong> Lowest odds lost
                  </span>
                </>
              )}
              {highestOddsWon !== null && (
                <>
                  <span className="text-border">|</span>
                  <span className="flex items-center gap-1">
                    <strong className="text-green-500">{highestOddsWon.toFixed(2)}</strong> Highest odds won
                  </span>
                </>
              )}
            </div>

            {/* Mobile-optimized stats display - initially hidden */}
            <div id="mobile-stats-cards" className="grid grid-cols-2 gap-3 mb-8 md:hidden hidden">
              <div className="bg-card border border-border/50 rounded-lg p-3 text-center">
                <div className="text-xl font-bold text-foreground">{stats.total_bets}</div>
                <div className="text-xs text-muted-foreground mt-1">bets</div>
              </div>
              <div className="bg-card border border-border/50 rounded-lg p-3 text-center">
                <div className="text-xl font-bold text-green-500">{stats.win_rate.toFixed(1)}%</div>
                <div className="text-xs text-muted-foreground mt-1">win rate</div>
              </div>
              <div className="bg-card border border-border/50 rounded-lg p-3 text-center">
                <div className="text-xl font-bold text-foreground">{stats.avg_odds.toFixed(2)}</div>
                <div className="text-xs text-muted-foreground mt-1">Avg odds</div>
              </div>
              <div className="bg-card border border-border/50 rounded-lg p-3 text-center">
                <div className="text-xl font-bold text-blue-500">{Number(player.money).toLocaleString()}</div>
                <div className="text-xs text-muted-foreground mt-1">Balance</div>
              </div>
              <div className="bg-card border border-border/50 rounded-lg p-3 text-center">
                <div className="text-xl font-bold text-red-500">{lowestOddsLost !== null ? lowestOddsLost.toFixed(2) : 'N/A'}</div>
                <div className="text-xs text-muted-foreground mt-1">Lowest odds lost</div>
              </div>
              <div className="bg-card border border-border/50 rounded-lg p-3 text-center">
                <div className="text-xl font-bold text-green-500">{highestOddsWon !== null ? highestOddsWon.toFixed(2) : 'N/A'}</div>
                <div className="text-xs text-muted-foreground mt-1">Highest odds won</div>
              </div>
            </div>

            <div className="flex flex-wrap items-center gap-x-8 gap-y-4 mb-8">
              <div className="flex flex-wrap gap-x-8 gap-y-4">
                <button
                  onClick={() => {
                    setStatusFilter("all");
                    setCurrentPage(1);
                  }}
                  className={`text-sm font-medium transition-colors hover:underline decoration-primary/30 underline-offset-4 ${
                    statusFilter === "all"
                      ? "text-primary underline"
                      : "text-muted-foreground hover:text-primary"
                  }`}
                >
                  All Bets
                </button>
                <button
                  onClick={() => {
                    setStatusFilter("pending");
                    setCurrentPage(1);
                  }}
                  className={`text-sm font-medium transition-colors hover:underline decoration-primary/30 underline-offset-4 ${
                    statusFilter === "pending"
                      ? "text-primary underline"
                      : "text-muted-foreground hover:text-primary"
                  }`}
                >
                  Pending
                </button>
                <button
                  onClick={() => {
                    setStatusFilter("won");
                    setCurrentPage(1);
                  }}
                  className={`text-sm font-medium transition-colors hover:underline decoration-primary/30 underline-offset-4 ${
                    statusFilter === "won"
                      ? "text-primary underline"
                      : "text-muted-foreground hover:text-primary"
                  }`}
                >
                  Won
                </button>
                <button
                  onClick={() => {
                    setStatusFilter("lost");
                    setCurrentPage(1);
                  }}
                  className={`text-sm font-medium transition-colors hover:underline decoration-primary/30 underline-offset-4 ${
                    statusFilter === "lost"
                      ? "text-primary underline"
                      : "text-muted-foreground hover:text-primary"
                  }`}
                >
                  Lost
                </button>
              </div>

              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={generatePDF}
                  className="gap-2 hover:bg-primary/10 hover:border-primary/50"
                >
                  <Download className="size-4" />
                  PDF
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={generateCSV}
                  className="gap-2 hover:bg-primary/10 hover:border-primary/50"
                >
                  <FileText className="size-4" />
                  CSV
                </Button>
              </div>
            </div>
          </motion.div>

          {processedBets.length === 0 ? (
            <div className="text-center py-16 text-muted-foreground bg-card border border-border/50 rounded-2xl">
              <p className="text-lg">No bets found</p>
            </div>
          ) : (
            <>
              {/* Desktop Table View */}
              <div className="hidden md:block bg-card border border-border/50 rounded-2xl overflow-hidden shadow-sm overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow className="hover:bg-transparent border-border/50">
                      <TableHead className="w-[60px] text-xs uppercase font-bold text-center">#</TableHead>
                      <TableHead className="cursor-pointer" onClick={() => handleSort('created_at')}>
                        <span className="flex items-center gap-1 text-xs uppercase font-bold">
                          Bet Date <ArrowUpDown className="size-3" />
                        </span>
                      </TableHead>
                      <TableHead className="text-xs uppercase font-bold">Match</TableHead>
                      <TableHead className="text-xs uppercase font-bold">Pick</TableHead>
                      <TableHead className="cursor-pointer" onClick={() => handleSort('bet_amount')}>
                        <span className="flex items-center gap-1 text-xs uppercase font-bold">
                          Stake <ArrowUpDown className="size-3" />
                        </span>
                      </TableHead>
                      <TableHead className="cursor-pointer" onClick={() => handleSort('odds')}>
                        <span className="flex items-center gap-1 text-xs uppercase font-bold">
                          Odds <ArrowUpDown className="size-3" />
                        </span>
                      </TableHead>
                      <TableHead className="cursor-pointer" onClick={() => handleSort('potentialWin')}>
                        <span className="flex items-center gap-1 text-xs uppercase font-bold">
                          Potential <ArrowUpDown className="size-3" />
                        </span>
                      </TableHead>
                      <TableHead className="text-xs uppercase font-bold">Status</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {paginatedBets.map((bet, index) => (
                      <TableRow key={bet.bet_id} className="border-border/50">
                        <TableCell className="text-center font-medium text-muted-foreground">
                          {(currentPage - 1) * itemsPerPage + index + 1}
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {format(new Date(bet.created_at), "MMM d, yyyy HH:mm")}
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-col">
                            <span className="font-medium text-foreground">
                              {bet.home_team} vs {bet.away_team}
                            </span>
                            <span className="text-xs text-muted-foreground mt-1">
                              {bet.commence_time ? format(new Date(bet.commence_time), "MMM d, yyyy HH:mm") : "N/A"}
                            </span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline" className="capitalize font-medium bg-purple-600 text-white border-purple-600 hover:bg-purple-700">
                            {bet.bet_type}
                          </Badge>
                        </TableCell>
                        <TableCell className="font-semibold text-foreground">
                          {Number(bet.bet_amount || 0).toLocaleString()}
                        </TableCell>
                        <TableCell className="text-foreground">
                          {Number(bet.odds || 0).toFixed(2)}
                        </TableCell>
                        <TableCell className="font-semibold text-primary">
                          {Number(bet.potential_win || 0).toLocaleString(undefined, { maximumFractionDigits: 2 })}
                        </TableCell>
                        <TableCell>{getStatusBadge(bet.status)}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>

              {/* Mobile Card View */}
              <div className="md:hidden grid grid-cols-1 gap-6">
                {paginatedBets.map((bet, index) => (
                  <BetCard
                    key={bet.bet_id}
                    bet={bet}
                    index={(currentPage - 1) * itemsPerPage + index}
                    totalBets={processedBets.length}
                  />
                ))}
              </div>

              {totalPages > 1 && (
                <div className="flex items-center justify-center gap-2 mt-8 py-4">
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
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
                        onClick={() => setCurrentPage(page)}
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
                    onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
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
  );
}
