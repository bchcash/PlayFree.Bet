import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Trophy } from "lucide-react";
import { useMutation } from "@tanstack/react-query";
import { apiRequest, queryClient } from "@/lib/queryClient";
import { useToast } from "@/hooks/use-toast";

interface BettingModalProps {
  open: boolean;
  onClose: () => void;
  matchId: string;
  homeTeam: string;
  awayTeam: string;
  betType: "home" | "draw" | "away";
  odds: number;
  userBalance: number;
  onSuccess?: () => void;
}

export function BettingModal({
  open,
  onClose,
  matchId,
  homeTeam,
  awayTeam,
  betType,
  odds,
  userBalance,
  onSuccess,
}: BettingModalProps) {
  const { toast } = useToast();
  const [betAmount, setBetAmount] = useState<string>("");

  useEffect(() => {
    if (open) {
      setBetAmount("");
    }
  }, [open]);

  const numericAmount = parseFloat(betAmount) || 0;
  const potentialWin = numericAmount * odds;
  const isValidAmount = numericAmount >= 10 && numericAmount <= userBalance;

  const mutation = useMutation({
    mutationFn: async () => {
      const response = await apiRequest("POST", "/api/bets", {
        match_id: matchId,
        bet_type: betType,
        odds,
        bet_amount: numericAmount,
        home_team: homeTeam,
        away_team: awayTeam,
      });
      return response.json();
    },
    onSuccess: (data) => {
      toast({
        title: "Bet Placed!",
        description: (
          <div>
            <div>Your bet of {numericAmount} has been placed.</div>
            <div>Potential win: {potentialWin.toFixed(2)}</div>
          </div>
        ),
      });
      queryClient.invalidateQueries({ queryKey: ["user"] });
      queryClient.invalidateQueries({ queryKey: ["user-bets"] });
      onClose();
      onSuccess?.();
    },
    onError: (error: Error) => {
      toast({
        title: "Error",
        description: error.message || "Failed to place bet. Please try again.",
        variant: "destructive",
      });
    },
  });

  const getBetTypeLabel = () => {
    switch (betType) {
      case "home":
        return `${homeTeam} Win`;
      case "draw":
        return "Draw";
      case "away":
        return `${awayTeam} Win`;
    }
  };

  const handleQuickAmount = (amount: number) => {
    if (amount <= userBalance) {
      setBetAmount(amount.toString());
    }
  };

  const handleSubmit = () => {
    if (isValidAmount) {
      mutation.mutate();
    }
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && onClose()}>
      <DialogContent className="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Trophy className="h-5 w-5" />
            Place Your Bet
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {/* <div className="bg-muted p-3 rounded-lg">
            <div className="text-sm text-muted-foreground mb-1">Match</div>
            <div className="font-medium">
              {homeTeam} vs {awayTeam}
            </div>
          </div> */}

          <div className="bg-muted p-3 rounded-lg">
            <div className="text-sm text-muted-foreground mb-1">Your Bet</div>
            <div className="text-lg">
              {getBetTypeLabel()} ({odds.toFixed(2)})
            </div>
          </div>

          <div className="bg-muted p-3 rounded-lg">
            <div className="flex justify-between items-center mb-3">
              <div className="text-sm text-muted-foreground">
                Quick Bet
              </div>
              <div className="text-sm text-muted-foreground">
                Balance: {userBalance.toLocaleString("en-US")}
              </div>
            </div>
            <div className="flex gap-2">
              {[100, 500, 1000].map((amount) => (
                <Button
                  key={amount}
                  variant="outline"
                  size="default"
                  onClick={() => handleQuickAmount(amount)}
                  disabled={amount > userBalance}
                  className="min-w-[60px] flex-1 bg-white text-black hover:bg-primary hover:text-primary-foreground border-white hover:border-primary text-base"
                >
                  {amount.toLocaleString("en-US")}
                </Button>
              ))}
              <Button
                variant="outline"
                size="default"
                onClick={() => handleQuickAmount(userBalance)}
                disabled={userBalance < 10}
                className="min-w-[60px] flex-1 bg-white text-black hover:bg-primary hover:text-primary-foreground border-white hover:border-primary font-bold text-base"
              >
                MAX
              </Button>
            </div>
          </div>

          <div className="bg-gray-100 dark:bg-gray-800 p-3 rounded-lg">
            <div className="text-sm text-muted-foreground mb-2">Bet Amount</div>
            <Input
              type="number"
              placeholder="Enter amount (min 10)"
              value={betAmount}
              onChange={(e) => setBetAmount(e.target.value)}
              min={10}
              max={userBalance}
              step={10}
              className="h-12 !text-lg placeholder:!text-lg"
            />
            {betAmount && !isValidAmount && (
              <div className="text-destructive text-sm mt-1">
                {numericAmount < 10
                  ? "Minimum bet is 10"
                  : numericAmount > userBalance
                    ? "Insufficient balance"
                    : "Invalid amount"}
              </div>
            )}
          </div>

          {isValidAmount && (
            <div className="bg-muted/70 p-3 rounded-lg">
              <div className="flex items-center justify-center gap-8">
                <div className="text-center">
                  <div className="text-lg font-bold text-foreground mb-1">
                    {Math.round(potentialWin).toLocaleString("en-US")}
                  </div>
                  <div className="text-sm text-muted-foreground">
                    Potential Win
                  </div>
                </div>
                <div className="text-center">
                  <div className="text-lg font-bold text-foreground mb-1">
                    {Math.round(numericAmount * (odds - 1)).toLocaleString("en-US")}
                  </div>
                  <div className="text-sm text-muted-foreground">
                    Profit
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        <div className="flex flex-col gap-3 mt-6">
          <Button
            onClick={handleSubmit}
            disabled={!isValidAmount || mutation.isPending}
            className="w-full h-12 bg-gray-900 hover:bg-gray-800 text-white border-gray-700 hover:border-gray-600 shadow-lg hover:shadow-xl transition-all duration-200 text-lg font-medium"
          >
            {mutation.isPending ? "Placing..." : "Place Bet"}
          </Button>
          <Button
            variant="outline"
            onClick={onClose}
            disabled={mutation.isPending}
            className="w-full h-12 border-gray-400 text-gray-600 hover:bg-gray-50 hover:border-gray-500 transition-all duration-200 text-lg"
          >
            Cancel
          </Button>
        </div>

        <div className="text-xs text-muted-foreground text-center pt-2 border-t">
          Virtual betting for practice only. No real money involved.
        </div>
      </DialogContent>
    </Dialog>
  );
}
