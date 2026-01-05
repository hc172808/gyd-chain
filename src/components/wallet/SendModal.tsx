import { useState } from "react";
import { ArrowUpRight, AlertCircle, Loader2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { toast } from "sonner";

interface SendModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  token: "GYDS" | "GYD";
  balance: number;
}

export function SendModal({ open, onOpenChange, token, balance }: SendModalProps) {
  const [recipient, setRecipient] = useState("");
  const [amount, setAmount] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const formatBalance = (balance: number) => {
    return new Intl.NumberFormat("en-US", {
      minimumFractionDigits: 2,
      maximumFractionDigits: 4,
    }).format(balance);
  };

  const handleSend = async () => {
    setError("");

    // Validation
    if (!recipient) {
      setError("Please enter a recipient address");
      return;
    }

    if (!recipient.startsWith("gyds1")) {
      setError("Invalid address format. Must start with 'gyds1'");
      return;
    }

    if (!amount || parseFloat(amount) <= 0) {
      setError("Please enter a valid amount");
      return;
    }

    if (parseFloat(amount) > balance) {
      setError("Insufficient balance");
      return;
    }

    setLoading(true);

    // Simulate transaction
    await new Promise((resolve) => setTimeout(resolve, 2000));

    setLoading(false);
    toast.success(`Successfully sent ${amount} ${token}`, {
      description: `To: ${recipient.slice(0, 12)}...${recipient.slice(-6)}`,
    });
    
    setRecipient("");
    setAmount("");
    onOpenChange(false);
  };

  const handleMax = () => {
    // Leave some for gas fees if GYDS
    const maxAmount = token === "GYDS" ? Math.max(0, balance - 0.01) : balance;
    setAmount(maxAmount.toString());
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <div className={`flex h-8 w-8 items-center justify-center rounded-lg ${token === "GYDS" ? "gradient-gyds" : "gradient-gyd"}`}>
              <ArrowUpRight className="h-4 w-4 text-white" />
            </div>
            Send {token}
          </DialogTitle>
          <DialogDescription>
            Transfer {token} tokens to another wallet address
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <div className="space-y-2">
            <Label htmlFor="recipient">Recipient Address</Label>
            <Input
              id="recipient"
              placeholder="gyds1..."
              value={recipient}
              onChange={(e) => setRecipient(e.target.value)}
              className="font-mono"
            />
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="amount">Amount</Label>
              <span className="text-xs text-muted-foreground">
                Available: {formatBalance(balance)} {token}
              </span>
            </div>
            <div className="relative">
              <Input
                id="amount"
                type="number"
                placeholder="0.00"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className="pr-16"
              />
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="absolute right-1 top-1/2 -translate-y-1/2 h-7 text-xs"
                onClick={handleMax}
              >
                MAX
              </Button>
            </div>
          </div>

          <div className="rounded-lg bg-muted p-4 space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-muted-foreground">Network Fee</span>
              <span>~0.001 GYDS</span>
            </div>
            <div className="flex justify-between text-sm font-medium">
              <span>Total</span>
              <span>
                {amount ? parseFloat(amount).toFixed(4) : "0.00"} {token}
              </span>
            </div>
          </div>
        </div>

        <div className="flex gap-3">
          <Button
            variant="outline"
            className="flex-1"
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button
            className={`flex-1 ${token === "GYDS" ? "gradient-gyds" : "gradient-gyd"} text-white border-0`}
            onClick={handleSend}
            disabled={loading}
          >
            {loading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Sending...
              </>
            ) : (
              <>
                <ArrowUpRight className="mr-2 h-4 w-4" />
                Send {token}
              </>
            )}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
