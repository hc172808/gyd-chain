import { ArrowUpRight, TrendingUp, DollarSign, Coins } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

interface BalanceCardsProps {
  gydsBalance: number;
  gydBalance: number;
  onSend: (token: "GYDS" | "GYD") => void;
}

export function BalanceCards({ gydsBalance, gydBalance, onSend }: BalanceCardsProps) {
  const formatBalance = (balance: number) => {
    return new Intl.NumberFormat("en-US", {
      minimumFractionDigits: 2,
      maximumFractionDigits: 4,
    }).format(balance);
  };

  // Mock price data
  const gydsPrice = 2.45;
  const gydPrice = 1.0; // Stablecoin
  const gydsChange = 5.23;

  return (
    <div className="grid gap-6 md:grid-cols-2">
      {/* GYDS Card */}
      <Card className="overflow-hidden border-0 shadow-xl">
        <div className="gradient-gyds p-6">
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-3">
              <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-white/20 backdrop-blur">
                <Coins className="h-6 w-6 text-white" />
              </div>
              <div>
                <p className="text-sm font-medium text-white/80">GYDS Token</p>
                <p className="text-xs text-white/60">Utility & Governance</p>
              </div>
            </div>
            <Button
              size="sm"
              variant="secondary"
              className="bg-white/20 text-white hover:bg-white/30 border-0"
              onClick={() => onSend("GYDS")}
            >
              <ArrowUpRight className="mr-1 h-4 w-4" />
              Send
            </Button>
          </div>

          <div className="mt-6">
            <p className="text-4xl font-bold text-white">
              {formatBalance(gydsBalance)}
            </p>
            <p className="mt-1 text-sm text-white/70">
              ≈ ${formatBalance(gydsBalance * gydsPrice)} USD
            </p>
          </div>

          <div className="mt-4 flex items-center gap-2">
            <div className="flex items-center gap-1 rounded-full bg-white/20 px-2 py-1">
              <TrendingUp className="h-3 w-3 text-white" />
              <span className="text-xs font-medium text-white">
                +{gydsChange}%
              </span>
            </div>
            <span className="text-xs text-white/60">24h</span>
          </div>
        </div>
      </Card>

      {/* GYD Card */}
      <Card className="overflow-hidden border-0 shadow-xl">
        <div className="gradient-gyd p-6">
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-3">
              <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-white/20 backdrop-blur">
                <DollarSign className="h-6 w-6 text-white" />
              </div>
              <div>
                <p className="text-sm font-medium text-white/80">GYD Stablecoin</p>
                <p className="text-xs text-white/60">Pegged to USD</p>
              </div>
            </div>
            <Button
              size="sm"
              variant="secondary"
              className="bg-white/20 text-white hover:bg-white/30 border-0"
              onClick={() => onSend("GYD")}
            >
              <ArrowUpRight className="mr-1 h-4 w-4" />
              Send
            </Button>
          </div>

          <div className="mt-6">
            <p className="text-4xl font-bold text-white">
              {formatBalance(gydBalance)}
            </p>
            <p className="mt-1 text-sm text-white/70">
              ≈ ${formatBalance(gydBalance * gydPrice)} USD
            </p>
          </div>

          <div className="mt-4 flex items-center gap-2">
            <div className="flex items-center gap-1 rounded-full bg-white/20 px-2 py-1">
              <DollarSign className="h-3 w-3 text-white" />
              <span className="text-xs font-medium text-white">$1.00</span>
            </div>
            <span className="text-xs text-white/60">Stable</span>
          </div>
        </div>
      </Card>
    </div>
  );
}
