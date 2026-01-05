import { Copy, ExternalLink, Shield } from "lucide-react";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";

interface WalletHeaderProps {
  address: string;
}

export function WalletHeader({ address }: WalletHeaderProps) {
  const copyAddress = () => {
    navigator.clipboard.writeText(address);
    toast.success("Address copied to clipboard");
  };

  const shortAddress = `${address.slice(0, 12)}...${address.slice(-8)}`;

  return (
    <header className="flex flex-col gap-6 md:flex-row md:items-center md:justify-between">
      <div className="flex items-center gap-4">
        <div className="flex h-14 w-14 items-center justify-center rounded-2xl gradient-gyds shadow-lg">
          <Shield className="h-7 w-7 text-primary-foreground" />
        </div>
        <div>
          <h1 className="text-2xl font-bold tracking-tight">GYDS Wallet</h1>
          <p className="text-sm text-muted-foreground">
            Secure blockchain wallet for GYDS & GYD tokens
          </p>
        </div>
      </div>

      <div className="flex items-center gap-2 rounded-xl bg-card px-4 py-3 shadow-sm border">
        <div className="h-2 w-2 rounded-full bg-green-500 animate-pulse" />
        <span className="font-mono text-sm">{shortAddress}</span>
        <Button
          variant="ghost"
          size="icon"
          className="h-8 w-8"
          onClick={copyAddress}
        >
          <Copy className="h-4 w-4" />
        </Button>
        <Button variant="ghost" size="icon" className="h-8 w-8">
          <ExternalLink className="h-4 w-4" />
        </Button>
      </div>
    </header>
  );
}
