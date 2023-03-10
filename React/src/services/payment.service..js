import axios from "axios";

//const PAYMENT_URL = "https://payment-4dcnj7fm6a-uc.a.run.app/api/"
//const PAYMENT_URL = "http://localhost:5054/api/"
const PAYMENT_URL = "http://34.122.120.6:5054/api/"

axios.defaults.withCredentials = true

let axiosConfig = {
    headers: {
        'Content-Type': 'text/plain',
       // "Access-Control-Allow-Origin": "http://localhost:3000"
    },
    withCredentials : true,
}

const makePayment = (payment) => {
    return axios.post(PAYMENT_URL + "payment", payment,
        axiosConfig)
            .then((response) => {
        return response;
      });
};

const PaymentService = {
    makePayment
}

  
export default PaymentService;

