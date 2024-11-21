from dataclasses import dataclass
import math

@dataclass
class GPUConfig:
    model: str
    price: float
    maintenance_cost_per_month: float
    users_per_gpu: int
    
    @staticmethod
    def calculate_maintenance_cost():
        # Power consumption and cost
        watts_per_hour = 450  # RTX 4090 power consumption
        hours_per_day = 24    # Assuming 24/7 operation
        days_per_month = 30
        kwh_rate = 0.037  # Kyrgyzstan  
        # kwh_rate = 0.188  # Mexico
        # kwh_rate = 0.188  # El Salvador
        # kwh_rate = 0.163  # Estonia
        # kwh_rate = 0.164  # Latvia
        # kwh_rate = 0.166  # Lithuania
        # kwh_rate = 0.070  # Kazakhstan
        # kwh_rate = 0.100  # Georgia
        # kwh_rate = 0.112  # Armenia
        # kwh_rate = 0.070  # Uzbekistan
        # kwh_rate = 0.089  # Russia
        # kwh_rate = 0.250  # Singapore
        # kwh_rate = 0.250  # Switzerland
        # kwh_rate = 0.361  # Germany
        # kwh_rate = 0.157  # United States
        # kwh_rate = 0.110  # Canada
        # kwh_rate = 0.421  # Netherlands
        # kwh_rate = 0.259  # Australia
        # kwh_rate = 0.223  # Sweden
        # kwh_rate = 0.275  # New Zealand
 
        
        # Monthly power cost
        monthly_power_cost = (watts_per_hour * hours_per_day * days_per_month / 1000) * kwh_rate

        # Cooling costs (typically 30-40% of power consumption)
        cooling_cost = monthly_power_cost * 0.35

        # Maintenance labor (periodic cleaning, monitoring, updates)
        labor_cost = 10  # Estimated monthly per-GPU labor cost

        # Parts replacement fund (setting aside money for potential repairs/replacements)
        parts_fund = 5   # Monthly allocation for future repairs

        total_cost = monthly_power_cost + cooling_cost + labor_cost + parts_fund
        return total_cost


@dataclass
class APIConfig:
    name: str
    cost_per_million_tokens: float
    tokens_per_minute: int  # Average tokens used per minute of usage

class BusinessCalculator:
    def __init__(
        self,
        total_users: int,
        paid_user_percentage: float,
        include_tts: bool,
        paid_hours_per_month: int = 30,  # Changed from 10 to 30 hours per month for paid users
        free_hours_per_month: int = 5,   # Changed from 2 to 5 hours per month for free users
        tax_of_country: float = 0.10    # Default tax rate for Kyrgyzstan
    ):
        # Calculate maintenance cost
        maintenance_cost = GPUConfig.calculate_maintenance_cost()
        
        # Fixed configurations
        self.subscription_price = 20.0  # Fixed $20 subscription
        self.tax_rate = tax_of_country  # Tax rate
        
        # GPU Configuration (RTX 4090 as default)
        self.gpu_config = GPUConfig(
            model="NVIDIA RTX 4090",
            price=1600.0,
            maintenance_cost_per_month=maintenance_cost,
            users_per_gpu=100
        )
        
                
        # API Configuration
        self.api_config = APIConfig(
            # GPT-4o-mini: A cost-efficient small model with vision capabilities. 
            # Smarter and cheaper than GPT-3.5 Turbo, with a 128K context limit and an October 2023 knowledge cutoff.
            name="gpt-4o-mini",
            cost_per_million_tokens=0.15,


            # GPT-4o: Our most advanced multimodal model, faster and cheaper than GPT-4 Turbo, 
            # with stronger vision capabilities. It has a 128K context limit and an October 2023 knowledge cutoff.
            # name="gpt-4o",
            # cost_per_million_tokens=2.50,

            # o1-mini: A fast, cost-efficient reasoning model designed for coding, math, and science. 
            # Offers a 128K context limit with a focus on high-speed processing.
            # name="o1-mini",
            # cost_per_million_tokens=3.0,


            # Fine-tuning models: Allows the creation of custom models by fine-tuning base models with your training data. 
            # Billing is based on tokens used in requests to the fine-tuned model.
            # Example: GPT-4o-2024-08-06, designed for custom fine-tuning.
            # name="gpt-4o-2024-08-06",
            # cost_per_million_tokens=3.75,

            # o1-preview: A cutting-edge reasoning model for complex tasks. 
            # Provides advanced reasoning capabilities with a 128K context limit and October 2023 knowledge cutoff.
            # name="o1-preview",
            # cost_per_million_tokens=15.0,

            

            # Average tokens processed per minute during a conversation.
            tokens_per_minute=500
        )

        # TTS Configuration
        self.tts_api_config = APIConfig(
            name="Text-to-Speech",
            cost_per_million_tokens=2.0,
            tokens_per_minute=100  # Example: 100 tokens per minute
        )
        
        # User inputs
        self.total_users = total_users
        self.paid_user_percentage = paid_user_percentage
        self.include_tts = include_tts
        self.paid_hours_per_month = paid_hours_per_month
        self.free_hours_per_month = free_hours_per_month
        
    @property
    def paid_users(self):
        return int((self.paid_user_percentage / 100) * self.total_users)
    
    @property
    def free_users(self):
        return self.total_users - self.paid_users
        
    def calculate_gpu_needs(self):
        return math.ceil(self.total_users / self.gpu_config.users_per_gpu)
        
    def calculate_initial_gpu_cost(self):
        return self.calculate_gpu_needs() * self.gpu_config.price
        
    def calculate_monthly_gpu_maintenance(self):
        return self.calculate_gpu_needs() * self.gpu_config.maintenance_cost_per_month
        
    def calculate_monthly_api_cost(self):
        # Calculate total tokens for paid users
        paid_tokens = (
            self.paid_users * 
            self.paid_hours_per_month * 
            60 * 
            self.api_config.tokens_per_minute
        )
        
        # Calculate total tokens for free users
        free_tokens = (
            self.free_users * 
            self.free_hours_per_month * 
            60 * 
            self.api_config.tokens_per_minute
        )
        
        total_tokens = paid_tokens + free_tokens
        return (total_tokens / 1_000_000) * self.api_config.cost_per_million_tokens
        
    def calculate_monthly_tts_cost(self):
        if not self.include_tts:
            return 0
            
        # Calculate tokens for Text-to-Speech
        paid_tokens = (
            self.paid_users * 
            self.paid_hours_per_month * 
            60 * 
            self.tts_api_config.tokens_per_minute
        )
        
        free_tokens = (
            self.free_users * 
            self.free_hours_per_month * 
            60 * 
            self.tts_api_config.tokens_per_minute
        )
        
        total_tokens = paid_tokens + free_tokens
        return (total_tokens / 1_000_000) * self.tts_api_config.cost_per_million_tokens
        
    def calculate_monthly_revenue(self):
        return self.paid_users * self.subscription_price
        
    def generate_report(self):
        gpus_needed = self.calculate_gpu_needs()
        initial_gpu_cost = self.calculate_initial_gpu_cost()
        monthly_gpu_maintenance = self.calculate_monthly_gpu_maintenance()
        monthly_api_cost = self.calculate_monthly_api_cost()
        monthly_tts_cost = self.calculate_monthly_tts_cost()
        monthly_revenue = self.calculate_monthly_revenue()
        
        # Include tax in monthly costs
        total_monthly_cost = monthly_gpu_maintenance + monthly_api_cost + monthly_tts_cost
        total_monthly_cost_with_tax = total_monthly_cost * (1 + self.tax_rate)
        monthly_profit = monthly_revenue - total_monthly_cost_with_tax
        
        return {
            "users": {
                "total": self.total_users,
                "paid": self.paid_users,
                "free": self.free_users,
                "paid_hours_per_month": self.paid_hours_per_month,
                "free_hours_per_month": self.free_hours_per_month
            },
            "gpu": {
                "model": self.gpu_config.model,
                "units_needed": gpus_needed,
                "initial_investment": initial_gpu_cost
            },
            "monthly_costs": {
                "gpu_maintenance": monthly_gpu_maintenance,
                "api_costs": monthly_api_cost,
                "tts_costs": monthly_tts_cost,
                "total_with_tax": total_monthly_cost_with_tax
            },
            "monthly_revenue": monthly_revenue,
            "monthly_profit": monthly_profit
        }

def main():
    # Get user inputs
    total_users = int(input("Enter total number of users: "))
    paid_percentage = float(input("Enter percentage of paid users (e.g., 10 for 10%): "))
    include_tts = input("Include Text-to-Speech? (yes/no): ").lower() == 'yes'
    
    # Tax rates for different countries - uncomment the one you want to use
    tax_of_country = 0.10  # Kyrgyzstan
    # tax_of_country = 0.16  # Mexico
    # tax_of_country = 0.13  # El Salvador
    # tax_of_country = 0.20  # Estonia
    # tax_of_country = 0.21  # Latvia
    # tax_of_country = 0.21  # Lithuania
    # tax_of_country = 0.12  # Kazakhstan
    # tax_of_country = 0.18  # Georgia
    # tax_of_country = 0.20  # Armenia
    # tax_of_country = 0.15  # Uzbekistan
    # tax_of_country = 0.20  # Russia
    # tax_of_country = 0.08  # Singapore
    # tax_of_country = 0.07  # Switzerland
    # tax_of_country = 0.19  # Germany
    # tax_of_country = 0.21  # United States
    # tax_of_country = 0.13  # Canada
    # tax_of_country = 0.21  # Netherlands
    # tax_of_country = 0.10  # Australia
    # tax_of_country = 0.25  # Sweden
    # tax_of_country = 0.15  # New Zealand
    
    # Create calculator with user inputs
    calculator = BusinessCalculator(
        total_users=total_users,
        paid_user_percentage=paid_percentage,
        include_tts=include_tts,
        tax_of_country=tax_of_country  # Use the selected tax rate
    )
    # Generate and print report
    report = calculator.generate_report()
    
    print("\n=== Business Metrics Report ===")
    print(f"Total Users: {report['users']['total']:,}")
    print(f"- Paid Users: {report['users']['paid']:,} ({report['users']['paid_hours_per_month']} hours/month)")
    print(f"- Free Users: {report['users']['free']:,} ({report['users']['free_hours_per_month']} hours/month)")
    
    print(f"\nGPU Requirements:")
    print(f"Model: {report['gpu']['model']}")
    print(f"Units Needed: {report['gpu']['units_needed']}")
    print(f"Initial Investment: ${report['gpu']['initial_investment']:,.2f}")
    
    print("\nMonthly Costs (with Tax):")
    print(f"GPU Maintenance: ${report['monthly_costs']['gpu_maintenance']:,.2f}")
    print(f"API Costs: ${report['monthly_costs']['api_costs']:,.2f}")
    print(f"TTS Costs: ${report['monthly_costs']['tts_costs']:,.2f}")
    print(f"Total Monthly Costs (with Tax): ${report['monthly_costs']['total_with_tax']:,.2f}")
    
    print(f"\nMonthly Revenue: ${report['monthly_revenue']:,.2f}")
    print(f"Monthly Profit: ${report['monthly_profit']:,.2f}")

if __name__ == "__main__":
    main()
