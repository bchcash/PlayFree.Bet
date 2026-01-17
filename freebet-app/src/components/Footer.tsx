import { Link } from "wouter";
import { Mail, MessageCircle, Globe, Twitter, Github } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
// import { useState, useEffect } from "react";

export function Footer() {
  // const [visitorCount, setVisitorCount] = useState<number | null>(null);

  // useEffect(() => {
  //   const fetchVisitorCount = async () => {
  //     try {
  //       const response = await fetch('/api/analytics');
  //       if (response.ok) {
  //         const data = await response.json();
  //         setVisitorCount(data.total_visitors || 0);
  //       }
  //     } catch (error) {
  //       console.warn('Failed to fetch visitor count:', error);
  //       setVisitorCount(0);
  //     }
  //   };

  //   fetchVisitorCount();
  //   // Update every 5 minutes
  //   const interval = setInterval(fetchVisitorCount, 5 * 60 * 1000);
  //   return () => clearInterval(interval);
  // }, []);

  return (
    <footer className="border-t bg-background py-12 mt-auto">
      <div className="container mx-auto px-4">
        <div className="flex flex-col items-center text-center">
          <div className="flex items-center gap-2 mb-6">
            <span className="font-display font-black text-xl tracking-tighter italic">
              FREEBET<span className="text-primary not-italic">GURU</span>
            </span>
          </div>

          <div className="flex flex-wrap justify-center gap-x-8 gap-y-4 mb-8 text-sm font-medium text-muted-foreground">
            <Link href="/" className="hover:text-primary transition-colors hover:underline decoration-primary/30 underline-offset-4">Matches</Link>
            <Link href="/leaderboard" className="hover:text-primary transition-colors hover:underline decoration-primary/30 underline-offset-4">Leaderboard</Link>
          </div>

          <div className="flex items-center gap-4 mb-10">
            <a href="https://t.me/freebet_guru" className="w-10 h-10 rounded-full bg-muted flex items-center justify-center text-muted-foreground hover:bg-primary/10 hover:text-primary transition-all">
              <MessageCircle className="size-5" />
            </a>
            <a href="mailto:freebet.guru@proton.me" className="w-10 h-10 rounded-full bg-muted flex items-center justify-center text-muted-foreground hover:bg-primary/10 hover:text-primary transition-all">
              <Mail className="size-5" />
            </a>
            <a href="https://x.com/freebet_guru" target="_blank" rel="noopener noreferrer" className="w-10 h-10 rounded-full bg-muted flex items-center justify-center text-muted-foreground hover:bg-primary/10 hover:text-primary transition-all">
              <Twitter className="size-5" />
            </a>
            <a href="https://github.com/fcbk1954/freebet.guru" target="_blank" rel="noopener noreferrer" className="w-10 h-10 rounded-full bg-muted flex items-center justify-center text-muted-foreground hover:bg-primary/10 hover:text-primary transition-all">
              <Github className="size-5" />
            </a>
            
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" className="h-10 px-4 rounded-full bg-muted hover:bg-primary/10 hover:text-primary text-muted-foreground transition-all gap-2 border-none">
                  <Globe className="size-4" />
                  <span className="text-xs font-bold uppercase tracking-wider">English</span>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="center" className="rounded-xl border shadow-xl p-1 w-32">
                <DropdownMenuItem className="rounded-lg font-bold text-xs uppercase tracking-wider focus:bg-primary/10 focus:text-primary">
                  English
                </DropdownMenuItem>
                <DropdownMenuItem disabled className="rounded-lg font-bold text-xs uppercase tracking-wider opacity-50 cursor-not-allowed">
                  Russian
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>

          {/* Visitor Counter - Commented out until Cloudflare API is configured */}
          {/* <div className="mb-4 flex items-center justify-center gap-2 text-sm text-muted-foreground">
            <Users className="size-4" />
            <span>
              {visitorCount !== null ? (
                <>
                  <span className="font-medium text-foreground">{visitorCount.toLocaleString()}</span> visitors
                </>
              ) : (
                <span>Loading...</span>
              )}
            </span>
          </div> */}

          <div className="pt-4 border-t border-border/50 w-full max-w-xs mx-auto">
            <p className="text-[10px] font-bold uppercase tracking-[0.2em] text-muted-foreground opacity-40">
              © {new Date().getFullYear()} FREEBETGURU. All rights reserved.
            </p>
          </div>
        </div>
      </div>
    </footer>
  );
}
