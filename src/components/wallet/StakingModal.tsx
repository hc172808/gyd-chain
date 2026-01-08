import { useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Landmark, TrendingUp, Clock, Users, ChevronRight, Loader2 } from "lucide-react";
import { toast } from "sonner";

interface StakingModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  gydsBalance: number;
}

interface Validator {
  id: string;
  name: string;
  commission: number;
  totalStaked: number;
  apr: number;
  uptime: number;
}

const mockValidators: Validator[] = [
  { id: "1", name: "GYDS Foundation", commission: 5, totalStaked: 2500000, apr: 12.5, uptime: 99.9 },
  { id: "2", name: "CryptoStake Pro", commission: 8, totalStaked: 1800000, apr: 11.8, uptime: 99.5 },
  { id: "3", name: "BlockNode Labs", commission: 6, totalStaked: 1200000, apr: 12.2, uptime: 99.7 },
  { id: "4", name: "StakePool Alpha", commission: 10, totalStaked: 900000, apr: 11.2, uptime: 99.3 },
];

const mockStakedAmount = 5000;
const mockRewards = 125.75;
const mockUnbondingPeriod = 21;

export function StakingModal({ open, onOpenChange, gydsBalance }: StakingModalProps) {
  const [amount, setAmount] = useState("");
  const [selectedValidator, setSelectedValidator] = useState<Validator | null>(null);
  const [loading, setLoading] = useState(false);
  const [unstakeAmount, setUnstakeAmount] = useState("");

  const handleStake = async () => {
    if (!amount || !selectedValidator) {
      toast.error("Please enter amount and select a validator");
      return;
    }
    
    const stakeAmount = parseFloat(amount);
    if (stakeAmount <= 0 || stakeAmount > gydsBalance) {
      toast.error("Invalid stake amount");
      return;
    }

    setLoading(true);
    // Simulate staking transaction
    await new Promise((resolve) => setTimeout(resolve, 2000));
    setLoading(false);
    toast.success(`Successfully staked ${stakeAmount} GYDS with ${selectedValidator.name}`);
    setAmount("");
    setSelectedValidator(null);
  };

  const handleUnstake = async () => {
    if (!unstakeAmount) {
      toast.error("Please enter amount to unstake");
      return;
    }
    
    const amount = parseFloat(unstakeAmount);
    if (amount <= 0 || amount > mockStakedAmount) {
      toast.error("Invalid unstake amount");
      return;
    }

    setLoading(true);
    await new Promise((resolve) => setTimeout(resolve, 2000));
    setLoading(false);
    toast.success(`Unstaking ${amount} GYDS. Unbonding period: ${mockUnbondingPeriod} days`);
    setUnstakeAmount("");
  };

  const handleClaimRewards = async () => {
    setLoading(true);
    await new Promise((resolve) => setTimeout(resolve, 1500));
    setLoading(false);
    toast.success(`Claimed ${mockRewards} GYDS rewards!`);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2 text-xl">
            <Landmark className="h-5 w-5 text-primary" />
            Staking Dashboard
          </DialogTitle>
        </DialogHeader>

        {/* Staking Overview Cards */}
        <div className="grid grid-cols-3 gap-3 mt-4">
          <Card className="bg-gradient-to-br from-primary/10 to-primary/5">
            <CardContent className="p-4">
              <p className="text-xs text-muted-foreground">Total Staked</p>
              <p className="text-lg font-bold">{mockStakedAmount.toLocaleString()} GYDS</p>
            </CardContent>
          </Card>
          <Card className="bg-gradient-to-br from-green-500/10 to-green-500/5">
            <CardContent className="p-4">
              <p className="text-xs text-muted-foreground">Rewards</p>
              <p className="text-lg font-bold text-green-600">{mockRewards} GYDS</p>
            </CardContent>
          </Card>
          <Card className="bg-gradient-to-br from-orange-500/10 to-orange-500/5">
            <CardContent className="p-4">
              <p className="text-xs text-muted-foreground">Est. APR</p>
              <p className="text-lg font-bold text-orange-600">12.5%</p>
            </CardContent>
          </Card>
        </div>

        <Button 
          onClick={handleClaimRewards} 
          disabled={loading || mockRewards <= 0}
          className="w-full gradient-gyds mt-2"
        >
          {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <TrendingUp className="mr-2 h-4 w-4" />}
          Claim {mockRewards} GYDS Rewards
        </Button>

        <Tabs defaultValue="stake" className="mt-4">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="stake">Stake</TabsTrigger>
            <TabsTrigger value="unstake">Unstake</TabsTrigger>
          </TabsList>

          <TabsContent value="stake" className="space-y-4 mt-4">
            <div className="space-y-2">
              <Label>Amount to Stake</Label>
              <div className="relative">
                <Input
                  type="number"
                  placeholder="0.00"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  className="pr-20"
                />
                <div className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-2">
                  <button
                    onClick={() => setAmount(gydsBalance.toString())}
                    className="text-xs text-primary hover:underline"
                  >
                    MAX
                  </button>
                  <span className="text-sm font-medium">GYDS</span>
                </div>
              </div>
              <p className="text-xs text-muted-foreground">
                Available: {gydsBalance.toLocaleString()} GYDS
              </p>
            </div>

            <div className="space-y-2">
              <Label>Select Validator</Label>
              <div className="space-y-2 max-h-48 overflow-y-auto">
                {mockValidators.map((validator) => (
                  <button
                    key={validator.id}
                    onClick={() => setSelectedValidator(validator)}
                    className={`w-full p-3 rounded-lg border text-left transition-all ${
                      selectedValidator?.id === validator.id
                        ? "border-primary bg-primary/5"
                        : "border-border hover:border-primary/50"
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <div className="h-8 w-8 rounded-full bg-primary/20 flex items-center justify-center">
                          <Users className="h-4 w-4 text-primary" />
                        </div>
                        <div>
                          <p className="font-medium text-sm">{validator.name}</p>
                          <p className="text-xs text-muted-foreground">
                            {(validator.totalStaked / 1000000).toFixed(1)}M GYDS staked
                          </p>
                        </div>
                      </div>
                      <div className="text-right">
                        <Badge variant="secondary" className="text-green-600">
                          {validator.apr}% APR
                        </Badge>
                        <p className="text-xs text-muted-foreground mt-1">
                          {validator.commission}% fee
                        </p>
                      </div>
                    </div>
                    <div className="mt-2">
                      <div className="flex items-center justify-between text-xs mb-1">
                        <span className="text-muted-foreground">Uptime</span>
                        <span>{validator.uptime}%</span>
                      </div>
                      <Progress value={validator.uptime} className="h-1" />
                    </div>
                  </button>
                ))}
              </div>
            </div>

            <Button 
              onClick={handleStake} 
              disabled={loading || !amount || !selectedValidator}
              className="w-full gradient-gyds"
            >
              {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              Stake GYDS
              <ChevronRight className="ml-2 h-4 w-4" />
            </Button>
          </TabsContent>

          <TabsContent value="unstake" className="space-y-4 mt-4">
            <Card className="bg-orange-500/10 border-orange-500/20">
              <CardContent className="p-4 flex items-start gap-3">
                <Clock className="h-5 w-5 text-orange-500 mt-0.5" />
                <div>
                  <p className="font-medium text-sm">Unbonding Period</p>
                  <p className="text-xs text-muted-foreground">
                    Unstaking takes {mockUnbondingPeriod} days. During this period, you won't earn rewards.
                  </p>
                </div>
              </CardContent>
            </Card>

            <div className="space-y-2">
              <Label>Amount to Unstake</Label>
              <div className="relative">
                <Input
                  type="number"
                  placeholder="0.00"
                  value={unstakeAmount}
                  onChange={(e) => setUnstakeAmount(e.target.value)}
                  className="pr-20"
                />
                <div className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-2">
                  <button
                    onClick={() => setUnstakeAmount(mockStakedAmount.toString())}
                    className="text-xs text-primary hover:underline"
                  >
                    MAX
                  </button>
                  <span className="text-sm font-medium">GYDS</span>
                </div>
              </div>
              <p className="text-xs text-muted-foreground">
                Currently staked: {mockStakedAmount.toLocaleString()} GYDS
              </p>
            </div>

            <Button 
              onClick={handleUnstake} 
              disabled={loading || !unstakeAmount}
              variant="outline"
              className="w-full"
            >
              {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              Unstake GYDS
            </Button>
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}
