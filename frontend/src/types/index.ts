export interface Money {
  amount: string;
  currency: string;
}

export interface User {
  id: number;
  email: string;
  display_name: string;
  avatar_url: string;
  preferred_language: string;
}

export interface Workspace {
  id: number;
  name: string;
  description: string;
  owner_id: number;
  member_count: number;
  created_at: string;
  updated_at: string;
}

export interface InvitePreview {
  workspace_name: string;
  inviter_name: string;
  role: string;
  expires_at: string;
  expired: boolean;
  used: boolean;
}

export interface WorkspaceInvite {
  id: number;
  workspace_id: number;
  code: string;
  role: string;
  expires_at: string;
  created_at: string;
}

export interface MemberWithProfile {
  user_id: number;
  display_name: string;
  email: string;
  avatar_url: string;
  role: string;
  created_at: string;
}

export interface Category {
  id: number;
  workspace_id: number;
  name: string;
  icon: string;
  slug: string | null;
  created_at: string;
  updated_at: string;
}

export type TransactionSource = "manual" | "import" | "bank";
export type TransactionType = "expense" | "income" | "transfer";

export interface TransactionEntry {
  id: number;
  category_id: number;
  participant_id: number | null;
  amount: Money;
  note: string;
}

export interface Transaction {
  id: number;
  workspace_id: number;
  created_by: number;
  description: string;
  date: string;
  total_amount: Money;
  source: TransactionSource;
  type: TransactionType;
  notes: string;
  bank_name: string;
  iban: string | null;
  counterparty_iban: string | null;
  bank_account_id: number | null;
  entries: TransactionEntry[];
  created_at: string;
  updated_at: string;
}

export interface Funding {
  id: number;
  workspace_id: number;
  user_id: number;
  category_id: number | null;
  year_month: string;
  amount: Money;
  created_at: string;
  updated_at: string;
}

export type RuleScope = "system" | "user" | "workspace";

export interface Rule {
  id: number;
  scope: RuleScope;
  owner_id: number | null;
  workspace_id: number | null;
  match_pattern: string;
  target_category_id: number;
  priority: number;
  enabled: boolean;
  amount_min: string | null;
  amount_max: string | null;
  bank_account_id: number | null;
  counterparty_iban: string | null;
  created_at: string;
  updated_at: string;
}

export type ExpiryStatus = "ok" | "warning" | "expired";

export interface Institution {
  id: string;
  name: string;
  bic: string;
  logo: string;
  countries: string[];
}

export interface BankConnection {
  id: number;
  institution_id: string;
  institution_name: string;
  status: string;
  auth_expires_at: string;
  days_remaining: number;
  expiry_status: ExpiryStatus;
  created_at: string;
  updated_at: string;
}

export interface BankAccount {
  id: number;
  bank_connection_id: number;
  institution_name: string;
  external_id: string;
  iban: string | null;
  currency: string;
  name: string;
  custom_name: string;
  display_name: string;
  last_synced_at: string | null;
}

export interface CategorySpending {
  category_id: number;
  name: string;
  icon: string;
  slug: string | null;
  spent: Money;
  prev_spent: Money;
  budget: Money | null;
}

export interface DailySpending {
  date: string;
  spent: Money;
}

export interface Summary {
  workspace_id: number;
  from: string;
  to: string;
  total_spent: Money;
  transaction_count: number;
  prev_total_spent: Money;
  prev_transaction_count: number;
  overall_budget: Money | null;
  by_category: CategorySpending[];
  daily_trend: DailySpending[];
}
