import { useState } from "react";
import { WalletHeader } from "./WalletHeader";
import { BalanceCards } from "./BalanceCards";
import { QuickActions } from "./QuickActions";
import { TransactionHistory } from "./TransactionHistory";
import { SendModal } from "./SendModal";
import { ReceiveModal } from "./ReceiveModal";

export interface WalletData {
  address: string;
  gydsBalance: number;
  gydBalance: number;
}

export interface Transaction {
  id: string;
  type: "send" | "receive";
  token: "GYDS" | "GYD";
  amount: number;
  address: string;
  timestamp: Date;
  status: "pending" | "confirmed" | "failed";
  txHash: string;
}

const mockWallet: WalletData = {
  address: "gyds1qxy2kgd3xjqn4zmny7v8pqr5c6wx8mn0yvhfj9",
  gydsBalance: 15847.5234,
  gydBalance: 2500.00,
};

const mockTransactions: Transaction[] = [
  {
    id: "1",
    type: "receive",
    token: "GYDS",
    amount: 500.0,
    address: "gyds1abc...def",
    timestamp: new Date(Date.now() - 1000 * 60 * 30),
    status: "confirmed",
    txHash: "0x1234...5678",
  },
  {
    id: "2",
    type: "send",
    token: "GYD",
    amount: 100.0,
    address: "gyds1xyz...789",
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 2),
    status: "confirmed",
    txHash: "0xabcd...efgh",
  },
  {
    id: "3",
    type: "receive",
    token: "GYDS",
    amount: 1250.75,
    address: "gyds1mno...pqr",
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 24),
    status: "confirmed",
    txHash: "0x9876...5432",
  },
  {
    id: "4",
    type: "send",
    token: "GYDS",
    amount: 75.25,
    address: "gyds1stu...vwx",
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 48),
    status: "pending",
    txHash: "0xijkl...mnop",
  },
  {
    id: "5",
    type: "receive",
    token: "GYD",
    amount: 500.0,
    address: "gyds1yz0...123",
    timestamp: new Date(Date.now() - 1000 * 60 * 60 * 72),
    status: "confirmed",
    txHash: "0xqrst...uvwx",
  },
];

export function WalletDashboard() {
  const [wallet] = useState<WalletData>(mockWallet);
  const [transactions] = useState<Transaction[]>(mockTransactions);
  const [sendModalOpen, setSendModalOpen] = useState(false);
  const [receiveModalOpen, setReceiveModalOpen] = useState(false);
  const [selectedToken, setSelectedToken] = useState<"GYDS" | "GYD">("GYDS");

  const handleSend = (token: "GYDS" | "GYD") => {
    setSelectedToken(token);
    setSendModalOpen(true);
  };

  const handleReceive = () => {
    setReceiveModalOpen(true);
  };

  return (
    <div className="min-h-screen bg-background pb-32">
      <div className="container mx-auto max-w-6xl px-4 py-6">
        <WalletHeader address={wallet.address} />
        
        <div className="mt-8 space-y-8">
          <BalanceCards
            gydsBalance={wallet.gydsBalance}
            gydBalance={wallet.gydBalance}
            onSend={handleSend}
          />
          
          <TransactionHistory transactions={transactions} />
        </div>
      </div>

      {/* Sticky Quick Actions at bottom */}
      <div className="fixed bottom-0 left-0 right-0 bg-background/95 backdrop-blur-sm border-t z-50">
        <div className="container mx-auto max-w-6xl px-4 py-4">
          <QuickActions
            onSend={() => handleSend("GYDS")}
            onReceive={handleReceive}
          />
        </div>
      </div>

      <SendModal
        open={sendModalOpen}
        onOpenChange={setSendModalOpen}
        token={selectedToken}
        balance={selectedToken === "GYDS" ? wallet.gydsBalance : wallet.gydBalance}
      />

      <ReceiveModal
        open={receiveModalOpen}
        onOpenChange={setReceiveModalOpen}
        address={wallet.address}
      />
    </div>
  );
}
