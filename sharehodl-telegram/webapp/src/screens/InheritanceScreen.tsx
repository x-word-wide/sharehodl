/**
 * Inheritance Screen - Asset transfer and inheritance planning
 */

import { useState } from 'react';
import { Shield, Clock, Users, FileText, Plus, AlertTriangle, Check } from 'lucide-react';

interface Beneficiary {
  id: string;
  name: string;
  address: string;
  allocation: number;
}

export function InheritanceScreen() {
  const [showAddBeneficiary, setShowAddBeneficiary] = useState(false);
  const [beneficiaries, setBeneficiaries] = useState<Beneficiary[]>([
    { id: '1', name: 'John Doe', address: 'hodl1abc...xyz', allocation: 50 },
    { id: '2', name: 'Jane Doe', address: 'hodl1def...uvw', allocation: 50 }
  ]);
  const tg = window.Telegram?.WebApp;

  const totalAllocation = beneficiaries.reduce((sum, b) => sum + b.allocation, 0);

  return (
    <div className="flex flex-col min-h-screen bg-dark-bg">
      {/* Header */}
      <div className="p-4">
        <h1 className="text-xl font-bold text-white mb-2">Inheritance Planning</h1>
        <p className="text-gray-400 text-sm mb-4">
          Secure asset transfer to your loved ones
        </p>

        {/* Status Card */}
        <div className="p-4 bg-dark-card rounded-xl mb-4">
          <div className="flex items-center gap-3 mb-3">
            <div className="w-10 h-10 rounded-full bg-accent-green/20 flex items-center justify-center">
              <Shield className="text-accent-green" size={20} />
            </div>
            <div>
              <p className="text-white font-medium">Plan Status</p>
              <p className="text-accent-green text-sm">Active</p>
            </div>
          </div>
          <div className="flex items-center gap-2 text-sm text-gray-400">
            <Clock size={14} />
            <span>Last check-in: 2 days ago</span>
          </div>
        </div>
      </div>

      {/* Plan Options */}
      <div className="px-4 mb-4">
        <h3 className="text-white font-semibold mb-3">Transfer Methods</h3>
        <div className="space-y-3">
          <PlanOptionCard
            icon={<Clock size={20} />}
            title="Dead Man's Switch"
            description="Auto-transfer if no activity for 365 days"
            isActive
            color="accent-orange"
          />
          <PlanOptionCard
            icon={<Users size={20} />}
            title="Multi-Signature Recovery"
            description="Requires 2 of 3 trusted contacts"
            color="primary"
          />
          <PlanOptionCard
            icon={<FileText size={20} />}
            title="Time-Locked Transfer"
            description="Scheduled for a specific date"
            color="accent-blue"
          />
        </div>
      </div>

      {/* Beneficiaries */}
      <div className="px-4 flex-1">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-white font-semibold">Beneficiaries</h3>
          <button
            onClick={() => setShowAddBeneficiary(true)}
            className="flex items-center gap-1 text-primary text-sm"
          >
            <Plus size={16} />
            Add
          </button>
        </div>

        {/* Allocation warning */}
        {totalAllocation !== 100 && (
          <div className="flex items-center gap-2 p-3 bg-accent-orange/10 border border-accent-orange/30 rounded-xl mb-3">
            <AlertTriangle className="text-accent-orange" size={18} />
            <span className="text-accent-orange text-sm">
              Total allocation must equal 100% (currently {totalAllocation}%)
            </span>
          </div>
        )}

        <div className="space-y-2 pb-24">
          {beneficiaries.map((beneficiary) => (
            <BeneficiaryCard
              key={beneficiary.id}
              beneficiary={beneficiary}
              onEdit={() => tg?.showAlert('Edit beneficiary')}
            />
          ))}
        </div>
      </div>

      {/* Add Beneficiary Dialog */}
      {showAddBeneficiary && (
        <AddBeneficiaryDialog
          onClose={() => setShowAddBeneficiary(false)}
          onAdd={(beneficiary) => {
            setBeneficiaries([...beneficiaries, { ...beneficiary, id: Date.now().toString() }]);
            setShowAddBeneficiary(false);
          }}
        />
      )}
    </div>
  );
}

function PlanOptionCard({
  icon,
  title,
  description,
  isActive,
  color
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
  isActive?: boolean;
  color: string;
}) {
  const tg = window.Telegram?.WebApp;

  return (
    <button
      onClick={() => tg?.showAlert(`Configure ${title}`)}
      className={`w-full p-4 rounded-xl flex items-center gap-3 ${
        isActive ? `bg-${color}/10 border border-${color}/30` : 'bg-dark-card'
      }`}
    >
      <div className={`w-10 h-10 rounded-full bg-${color}/20 flex items-center justify-center`}>
        <span className={`text-${color}`}>{icon}</span>
      </div>
      <div className="flex-1 text-left">
        <div className="flex items-center gap-2">
          <span className="text-white font-medium">{title}</span>
          {isActive && <Check size={14} className="text-accent-green" />}
        </div>
        <p className="text-gray-500 text-sm">{description}</p>
      </div>
    </button>
  );
}

function BeneficiaryCard({
  beneficiary,
  onEdit
}: {
  beneficiary: Beneficiary;
  onEdit: () => void;
}) {
  return (
    <div className="p-4 bg-dark-card rounded-xl">
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-full bg-primary/20 flex items-center justify-center">
            <span className="text-primary font-semibold">
              {beneficiary.name.split(' ').map(n => n[0]).join('')}
            </span>
          </div>
          <div>
            <p className="text-white font-medium">{beneficiary.name}</p>
            <p className="text-gray-500 text-sm">{beneficiary.address}</p>
          </div>
        </div>
        <div className="text-right">
          <p className="text-white font-semibold">{beneficiary.allocation}%</p>
          <p className="text-gray-500 text-xs">Allocation</p>
        </div>
      </div>
      <button
        onClick={onEdit}
        className="w-full mt-2 py-2 bg-dark-surface rounded-lg text-gray-400 text-sm"
      >
        Edit
      </button>
    </div>
  );
}

function AddBeneficiaryDialog({
  onClose,
  onAdd
}: {
  onClose: () => void;
  onAdd: (beneficiary: Omit<Beneficiary, 'id'>) => void;
}) {
  const [name, setName] = useState('');
  const [address, setAddress] = useState('');
  const [allocation, setAllocation] = useState('');

  return (
    <div className="fixed inset-0 bg-black/60 z-50 flex items-end">
      <div className="w-full bg-dark-card rounded-t-3xl p-4 animate-slide-up">
        <h3 className="text-white font-semibold text-lg mb-4">Add Beneficiary</h3>

        <div className="space-y-4 mb-6">
          <div>
            <label className="text-sm text-gray-400 mb-2 block">Name</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Full name"
              className="input"
            />
          </div>

          <div>
            <label className="text-sm text-gray-400 mb-2 block">Wallet Address</label>
            <input
              type="text"
              value={address}
              onChange={(e) => setAddress(e.target.value)}
              placeholder="hodl1..."
              className="input"
            />
          </div>

          <div>
            <label className="text-sm text-gray-400 mb-2 block">Allocation (%)</label>
            <input
              type="number"
              value={allocation}
              onChange={(e) => setAllocation(e.target.value)}
              placeholder="0"
              min="1"
              max="100"
              className="input"
            />
          </div>
        </div>

        <div className="flex gap-3">
          <button onClick={onClose} className="flex-1 btn-secondary">
            Cancel
          </button>
          <button
            onClick={() => onAdd({ name, address, allocation: parseInt(allocation) })}
            disabled={!name || !address || !allocation}
            className="flex-1 btn-primary"
          >
            Add Beneficiary
          </button>
        </div>
      </div>
    </div>
  );
}
