import { useState, useEffect } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { RefreshCw, ArrowDownUp, Info, Loader2, TrendingUp, TrendingDown } from "lucide-react";
import { toast } from "sonner";

interface SwapModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  gydsBalance: number;
  gydBalance: number;
}

// Mock exchange rate: 1 GYDS = 2.45 GYD
const MOCK_RATE = 2.45;
const MOCK_FEE_PERCENT = 0.3;

export function SwapModal({ open, onOpenChange, gydsBalance, gydBalance }: SwapModalProps) {
  const [fromToken, setFromToken] = useState<"GYDS" | "GYD">("GYDS");
  const [fromAmount, setFromAmount] = useState("");
  const [toAmount, setToAmount] = useState("");
  const [loading, setLoading] = useState(false);
  const [priceImpact, setPriceImpact] = useState(0);

  const toToken = fromToken === "GYDS" ? "GYD" : "GYDS";
  const fromBalance = fromToken === "GYDS" ? gydsBalance : gydBalance;
  const toBalance = fromToken === "GYDS" ? gydBalance : gydsBalance;
  const rate = fromToken === "GYDS" ? MOCK_RATE : 1 / MOCK_RATE;

  useEffect(() => {
    if (fromAmount) {
      const amount = parseFloat(fromAmount);
      if (!isNaN(amount)) {
        const output = amount * rate * (1 - MOCK_FEE_PERCENT / 100);
        setToAmount(output.toFixed(4));
        // Simulate price impact based on amount
        setPriceImpact(Math.min(amount / 10000, 2));
      }
    } else {
      setToAmount("");
      setPriceImpact(0);
    }
  }, [fromAmount, rate]);

  const handleSwapTokens = () => {
    setFromToken(fromToken === "GYDS" ? "GYD" : "GYDS");
    setFromAmount(toAmount);
  };

  const handleSwap = async () => {
    if (!fromAmount) {
      toast.error("Please enter an amount");
      return;
    }

    const amount = parseFloat(fromAmount);
    if (amount <= 0 || amount > fromBalance) {
      toast.error("Invalid amount");
      return;
    }

    setLoading(true);
    // Simulate swap transaction
    await new Promise((resolve) => setTimeout(resolve, 2000));
    setLoading(false);
    toast.success(`Swapped ${fromAmount} ${fromToken} for ${toAmount} ${toToken}`);
    setFromAmount("");
    onOpenChange(false);
  };

  const handleMaxClick = () => {
    setFromAmount(fromBalance.toString());
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-xl">
            <RefreshCw className="h-5 w-5 text-primary" />
            Swap Tokens
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4 mt-4">
          {/* From Token */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label>From</Label>
              <span className="text-xs text-muted-foreground">
                Balance: {fromBalance.toLocaleString()} {fromToken}
              </span>
            </div>
            <Card className={fromToken === "GYDS" ? "border-primary/30 bg-primary/5" : "border-purple-500/30 bg-purple-500/5"}>
              <CardContent className="p-4">
                <div className="flex items-center gap-3">
                  <div className={`h-10 w-10 rounded-full flex items-center justify-center ${
                    fromToken === "GYDS" ? "gradient-gyds" : "gradient-gyd"
                  }`}>
                    <span className="text-primary-foreground font-bold text-sm">
                      {fromToken === "GYDS" ? "G" : "$"}
                    </span>
                  </div>
                  <div className="flex-1">
                    <p className="font-semibold">{fromToken}</p>
                    <p className="text-xs text-muted-foreground">
                      {fromToken === "GYDS" ? "Utility & Governance" : "Stablecoin"}
                    </p>
                  </div>
                  <div className="text-right">
                    <Input
                      type="number"
                      placeholder="0.00"
                      value={fromAmount}
                      onChange={(e) => setFromAmount(e.target.value)}
                      className="w-32 text-right text-lg font-semibold border-0 bg-transparent p-0 h-auto focus-visible:ring-0"
                    />
                    <button
                      onClick={handleMaxClick}
                      className="text-xs text-primary hover:underline"
                    >
                      MAX
                    </button>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Swap Direction Button */}
          <div className="flex justify-center -my-2 relative z-10">
            <Button
              variant="outline"
              size="icon"
              className="rounded-full h-10 w-10 bg-background shadow-md hover:rotate-180 transition-transform duration-300"
              onClick={handleSwapTokens}
            >
              <ArrowDownUp className="h-4 w-4" />
            </Button>
          </div>

          {/* To Token */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label>To</Label>
              <span className="text-xs text-muted-foreground">
                Balance: {toBalance.toLocaleString()} {toToken}
              </span>
            </div>
            <Card className={toToken === "GYDS" ? "border-primary/30 bg-primary/5" : "border-purple-500/30 bg-purple-500/5"}>
              <CardContent className="p-4">
                <div className="flex items-center gap-3">
                  <div className={`h-10 w-10 rounded-full flex items-center justify-center ${
                    toToken === "GYDS" ? "gradient-gyds" : "gradient-gyd"
                  }`}>
                    <span className="text-primary-foreground font-bold text-sm">
                      {toToken === "GYDS" ? "G" : "$"}
                    </span>
                  </div>
                  <div className="flex-1">
                    <p className="font-semibold">{toToken}</p>
                    <p className="text-xs text-muted-foreground">
                      {toToken === "GYDS" ? "Utility & Governance" : "Stablecoin"}
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="text-lg font-semibold">{toAmount || "0.00"}</p>
                    <p className="text-xs text-muted-foreground">Estimated</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Swap Details */}
          <Card className="bg-muted/50">
            <CardContent className="p-4 space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground flex items-center gap-1">
                  <Info className="h-3 w-3" /> Rate
                </span>
                <span className="font-medium">
                  1 {fromToken} = {rate.toFixed(4)} {toToken}
                </span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">Fee ({MOCK_FEE_PERCENT}%)</span>
                <span className="font-medium">
                  {fromAmount ? (parseFloat(fromAmount) * MOCK_FEE_PERCENT / 100).toFixed(4) : "0"} {fromToken}
                </span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted-foreground">Price Impact</span>
                <span className={`font-medium flex items-center gap-1 ${
                  priceImpact > 1 ? "text-red-500" : priceImpact > 0.5 ? "text-orange-500" : "text-green-500"
                }`}>
                  {priceImpact > 0.5 ? <TrendingDown className="h-3 w-3" /> : <TrendingUp className="h-3 w-3" />}
                  {priceImpact.toFixed(2)}%
                </span>
              </div>
            </CardContent>
          </Card>

          <Button 
            onClick={handleSwap} 
            disabled={loading || !fromAmount}
            className="w-full gradient-gyds"
          >
            {loading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Swapping...
              </>
            ) : (
              <>
                <RefreshCw className="mr-2 h-4 w-4" />
                Swap {fromToken} for {toToken}
              </>
            )}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
