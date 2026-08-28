/**
 * Admin API Service
 * =================
 * HTTP client for communicating with the Regional Agent admin endpoints.
 */

const API_BASE_URL = import.meta.env.VITE_ADMIN_API_URL || 'http://localhost:8080';

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}

export interface BranchSummary {
  branch_id: string;
  today_sales: number;
  today_transactions: number;
  current_balance: number;
  active_sessions: number;
}

export interface InventoryStatus {
  product_id: string;
  branch_id: string;
  current_quantity: number;
  available_quantity: number;
  is_low_stock: boolean;
}

export interface HealthStatus {
  status: string;
  agent_id: string;
  region_id: string;
  agent_state: string;
  raft_state: string;
  timestamp: string;
  checks?: Record<string, string>;
}

class AdminApiService {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl.replace(/\/$/, '');
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    try {
      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
      });

      const data = await response.json();
      if (!response.ok) {
        return {
          success: false,
          error: data.error || `HTTP ${response.status}`,
        };
      }

      return {
        success: true,
        data,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Network error',
      };
    }
  }

  async getHealth(): Promise<ApiResponse<HealthStatus>> {
    return this.request<HealthStatus>('/health');
  }

  async getBranchSummary(branchId: string): Promise<ApiResponse<BranchSummary>> {
    return this.request<BranchSummary>(`/api/v1/branches/${encodeURIComponent(branchId)}`);
  }

  async getInventory(productId: string, branchId?: string): Promise<ApiResponse<InventoryStatus>> {
    const url = new URL(`${this.baseUrl}/api/v1/inventory/${encodeURIComponent(productId)}`);
    if (branchId) {
      url.searchParams.set('branch_id', branchId);
    }

    try {
      const response = await fetch(url.toString(), {
        headers: { 'Content-Type': 'application/json' },
      });

      const data = await response.json();
      if (!response.ok) {
        return {
          success: false,
          error: data.error || `HTTP ${response.status}`,
        };
      }

      return {
        success: true,
        data,
      };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Network error',
      };
    }
  }
}

export const adminApiService = new AdminApiService();
