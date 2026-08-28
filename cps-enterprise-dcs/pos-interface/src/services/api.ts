/**
 * POS API Service
 * ===============
 * HTTP client for communicating with the Local Agent.
 */

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}

export interface SessionStartResponse {
  session_id: string;
  cashier_id: string;
  register_id: string;
  opening_balance: number;
  started_at: string;
  status: string;
}

export interface SaleResponse {
  event_id: string;
  event_type: string;
  stream_id: string;
  created_at: string;
}

export interface BranchSummary {
  branch_id: string;
  agent_id: string;
  state: string;
  today_sales: number;
  sync_status: {
    last_sync?: string;
    pending_events: number;
    is_connected: boolean;
  };
}

class ApiService {
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
      return data;
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Network error',
      };
    }
  }

  async startSession(data: {
    cashier_id: string;
    register_id: string;
    opening_balance: number;
  }): Promise<ApiResponse<SessionStartResponse>> {
    return this.request<SessionStartResponse>('/api/v1/sessions', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async closeSession(
    sessionId: string,
    data: {
      closing_balance: number;
      total_sales: number;
      transaction_count: number;
    }
  ): Promise<ApiResponse<any>> {
    return this.request(`/api/v1/sessions/${sessionId}/close`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async recordSale(data: {
    product_id: string;
    quantity: number;
    unit_price: number;
    total_amount: number;
    cashier_id: string;
    session_id: string;
    customer_id?: string;
    payment_method?: string;
  }): Promise<ApiResponse<SaleResponse>> {
    return this.request<SaleResponse>('/api/v1/sales', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getBranchSummary(): Promise<ApiResponse<BranchSummary>> {
    return this.request<BranchSummary>('/api/v1/summary');
  }

  async getProducts(): Promise<ApiResponse<any[]>> {
    return this.request<any[]>('/api/v1/products');
  }

  async getInventory(productId: string): Promise<ApiResponse<any>> {
    return this.request(`/api/v1/inventory/${encodeURIComponent(productId)}`);
  }

  async healthCheck(): Promise<ApiResponse<any>> {
    return this.request('/health');
  }
}

export const apiService = new ApiService();
