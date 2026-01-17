import { Link } from "wouter";
import { Navigation } from "@/components/Navigation";
import { Footer } from "@/components/Footer";
import { Button } from "@/components/ui/button";
import { motion } from "framer-motion";
import { AlertTriangle, Home, Search, Trophy } from "lucide-react";
import { HelmetManager } from "@/components/HelmetManager";

export default function NotFound() {
  return (
    <>
      <HelmetManager
        title="Page Not Found - 404 Error | FreeBet Guru"
        description="Oops! The page you're looking for doesn't exist on FreeBet Guru. Return to our virtual football betting platform to place bets and compete with other players."
        keywords="404 error, page not found, freebet guru, virtual betting, football betting"
      />
      <div className="min-h-screen flex flex-col bg-background font-sans selection:bg-primary/20">
        <Navigation />

        <main className="flex-1 py-12 md:py-20">
          <div className="container mx-auto px-4 max-w-4xl">
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5 }}
              className="text-center"
            >
              {/* 404 Visual */}
              <motion.div
                initial={{ scale: 0.8, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                transition={{ delay: 0.2, duration: 0.5 }}
                className="mb-8"
              >
                <div className="relative inline-block">
                  <div className="text-8xl md:text-9xl font-black text-primary/10 select-none">404</div>
                  <div className="absolute inset-0 flex items-center justify-center">
                    <AlertTriangle className="size-16 md:size-20 text-primary animate-pulse" />
                  </div>
                </div>
              </motion.div>

              {/* Title */}
              <motion.h1
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.4, duration: 0.5 }}
                className="text-3xl md:text-4xl font-display font-bold tracking-tight text-foreground mb-4"
              >
                Page Not Found
              </motion.h1>

              {/* Description */}
              <motion.p
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.6, duration: 0.5 }}
                className="text-lg text-muted-foreground max-w-2xl mx-auto mb-8"
              >
                Looks like this page got a red card! The content you're looking for doesn't exist or may have been moved.
                Don't worry, there are plenty of other matches to bet on.
              </motion.p>

              {/* Action Buttons */}
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.8, duration: 0.5 }}
                className="flex flex-col sm:flex-row gap-4 justify-center items-center mb-12"
              >
                <Button
                  asChild
                  size="lg"
                  className="rounded-full px-8 bg-gradient-to-r from-primary to-primary/80 hover:from-primary/90 hover:to-primary shadow-lg hover:shadow-xl transition-all duration-300 transform hover:scale-[1.02] flex items-center gap-2"
                >
                  <Link href="/">
                    <Home className="size-5" />
                    Back to Home
                  </Link>
                </Button>

                <Button
                  asChild
                  variant="outline"
                  size="lg"
                  className="rounded-full px-8 border-primary/20 hover:bg-primary/5 flex items-center gap-2"
                >
                  <Link href="/">
                    <Search className="size-5" />
                    Browse Matches
                  </Link>
                </Button>

                <Button
                  asChild
                  variant="outline"
                  size="lg"
                  className="rounded-full px-8 border-primary/20 hover:bg-primary/5 flex items-center gap-2"
                >
                  <Link href="/leaderboard">
                    <Trophy className="size-5" />
                    View Leaderboard
                  </Link>
                </Button>
              </motion.div>

              {/* Fun Message */}
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ delay: 1.0, duration: 0.5 }}
                className="bg-primary/5 border border-primary/10 rounded-2xl p-6 max-w-md mx-auto"
              >
                <p className="text-sm text-muted-foreground">
                  💡 <strong>Pro tip:</strong> Try checking the URL for typos, or use the navigation menu above to find what you're looking for.
                </p>
              </motion.div>
            </motion.div>
          </div>
        </main>

        <Footer />
      </div>
    </>
  );
}
