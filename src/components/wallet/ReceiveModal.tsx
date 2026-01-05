import { Copy, Download, QrCode } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";

interface ReceiveModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  address: string;
}

export function ReceiveModal({ open, onOpenChange, address }: ReceiveModalProps) {
  const copyAddress = () => {
    navigator.clipboard.writeText(address);
    toast.success("Address copied to clipboard");
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg gradient-gyds">
              <QrCode className="h-4 w-4 text-white" />
            </div>
            Receive Tokens
          </DialogTitle>
          <DialogDescription>
            Share your wallet address to receive GYDS or GYD tokens
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* QR Code Placeholder */}
          <div className="flex justify-center">
            <div className="relative">
              <div className="h-48 w-48 rounded-2xl bg-white p-4 shadow-lg">
                <div className="h-full w-full rounded-xl bg-gradient-to-br from-muted to-muted/50 flex items-center justify-center">
                  <div className="grid grid-cols-5 gap-1">
                    {Array.from({ length: 25 }).map((_, i) => (
                      <div
                        key={i}
                        className={`h-6 w-6 rounded-sm ${
                          Math.random() > 0.5 ? "bg-foreground" : "bg-transparent"
                        }`}
                      />
                    ))}
                  </div>
                </div>
              </div>
              <div className="absolute -bottom-3 left-1/2 -translate-x-1/2">
                <div className="flex h-10 w-10 items-center justify-center rounded-full gradient-gyds shadow-lg">
                  <span className="text-xs font-bold text-white">G</span>
                </div>
              </div>
            </div>
          </div>

          {/* Address Display */}
          <div className="space-y-2">
            <p className="text-center text-sm text-muted-foreground">
              Your Wallet Address
            </p>
            <div className="flex items-center gap-2 rounded-xl bg-muted p-4">
              <code className="flex-1 text-center text-sm font-mono break-all">
                {address}
              </code>
            </div>
          </div>

          {/* Action Buttons */}
          <div className="flex gap-3">
            <Button variant="outline" className="flex-1" onClick={copyAddress}>
              <Copy className="mr-2 h-4 w-4" />
              Copy Address
            </Button>
            <Button variant="outline" className="flex-1">
              <Download className="mr-2 h-4 w-4" />
              Save QR
            </Button>
          </div>

          {/* Warning */}
          <div className="rounded-lg bg-yellow-50 dark:bg-yellow-900/20 p-4 border border-yellow-200 dark:border-yellow-800">
            <p className="text-sm text-yellow-800 dark:text-yellow-200">
              <strong>Important:</strong> Only send GYDS or GYD tokens to this address. 
              Sending other tokens may result in permanent loss.
            </p>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
