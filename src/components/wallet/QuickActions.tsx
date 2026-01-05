import { ArrowUpRight, ArrowDownLeft, RefreshCw, History, Landmark, Pickaxe } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

interface QuickActionsProps {
  onSend: () => void;
  onReceive: () => void;
}

export function QuickActions({ onSend, onReceive }: QuickActionsProps) {
  const actions = [
    {
      icon: ArrowUpRight,
      label: "Send",
      description: "Transfer tokens",
      onClick: onSend,
      gradient: "gradient-gyds",
    },
    {
      icon: ArrowDownLeft,
      label: "Receive",
      description: "Get your address",
      onClick: onReceive,
      gradient: "gradient-gyd",
    },
    {
      icon: RefreshCw,
      label: "Swap",
      description: "Exchange tokens",
      onClick: () => {},
      gradient: "bg-secondary",
    },
    {
      icon: Landmark,
      label: "Stake",
      description: "Earn rewards",
      onClick: () => {},
      gradient: "bg-secondary",
    },
    {
      icon: Pickaxe,
      label: "Mine",
      description: "Earn GYDS",
      onClick: () => {},
      gradient: "bg-secondary",
    },
    {
      icon: History,
      label: "History",
      description: "View all txns",
      onClick: () => {},
      gradient: "bg-secondary",
    },
  ];

  return (
    <Card className="border shadow-sm">
      <CardContent className="p-6">
        <h2 className="mb-4 text-lg font-semibold">Quick Actions</h2>
        <div className="grid grid-cols-3 gap-3 md:grid-cols-6">
          {actions.map((action) => (
            <Button
              key={action.label}
              variant="outline"
              className="flex h-auto flex-col gap-2 p-4 hover:shadow-md transition-all"
              onClick={action.onClick}
            >
              <div
                className={`flex h-10 w-10 items-center justify-center rounded-xl ${action.gradient}`}
              >
                <action.icon className="h-5 w-5 text-primary-foreground" />
              </div>
              <div className="text-center">
                <p className="text-sm font-medium">{action.label}</p>
                <p className="text-xs text-muted-foreground hidden md:block">
                  {action.description}
                </p>
              </div>
            </Button>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
