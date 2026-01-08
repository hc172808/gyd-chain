import { ArrowUpRight, ArrowDownLeft, RefreshCw, History, Landmark, Pickaxe } from "lucide-react";
import { Button } from "@/components/ui/button";

interface QuickActionsProps {
  onSend: () => void;
  onReceive: () => void;
  onStake: () => void;
  onSwap: () => void;
}

export function QuickActions({ onSend, onReceive, onStake, onSwap }: QuickActionsProps) {
  const actions = [
    {
      icon: ArrowUpRight,
      label: "Send",
      onClick: onSend,
      gradient: "gradient-gyds",
    },
    {
      icon: ArrowDownLeft,
      label: "Receive",
      onClick: onReceive,
      gradient: "gradient-gyd",
    },
    {
      icon: RefreshCw,
      label: "Swap",
      onClick: onSwap,
      gradient: "bg-gradient-to-br from-blue-500 to-cyan-500",
    },
    {
      icon: Landmark,
      label: "Stake",
      onClick: onStake,
      gradient: "bg-gradient-to-br from-orange-500 to-amber-500",
    },
    {
      icon: Pickaxe,
      label: "Mine",
      onClick: () => {},
      gradient: "bg-secondary",
    },
    {
      icon: History,
      label: "History",
      onClick: () => {},
      gradient: "bg-secondary",
    },
  ];

  return (
    <div className="grid grid-cols-3 gap-2 md:grid-cols-6 md:gap-3">
      {actions.map((action) => (
        <Button
          key={action.label}
          variant="outline"
          className="flex h-auto flex-col gap-1.5 p-3 hover:shadow-md transition-all"
          onClick={action.onClick}
        >
          <div
            className={`flex h-9 w-9 items-center justify-center rounded-xl ${action.gradient}`}
          >
            <action.icon className="h-4 w-4 text-primary-foreground" />
          </div>
          <p className="text-xs font-medium">{action.label}</p>
        </Button>
      ))}
    </div>
  );
}
