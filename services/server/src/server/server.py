import socket
import logger
import safe_socket
from protocol import read_message, encode_bet_ack, encode_winner, encode_end, Winner
from lottery import Bet, Lottery

class Server:
    def __init__(self, server_host: str, server_port: int) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.lottery = Lottery("bets.csv")

    def _handle_client(self, client_socket):
        action = "handle-client"
        message_amount = 0
        agency = None
        try:
            logger.info(action, logger.LogResult.in_progress)
            while True:
                data = read_message(client_socket)
                if data is None: 
                    break 
                agency = data[0].agency
                bets = [ Bet(bet.agency, bet.nombre, bet.apellido, bet.documento, bet.cumpleanos, bet.number) 
                    for bet in data ]
                
                self.lottery.store_bets(bets)

                message_amount += len(bets)
                safe_socket.send_all(client_socket, encode_bet_ack(len(bets)))

            for bet in self.lottery.load_bets(): 
                if bet.agency_id == agency and self.lottery.has_won(bet): 
                    safe_socket.send_all(
                        client_socket,
                        encode_winner(Winner(
                            nombre=bet.first_name,
                            apellido=bet.last_name,
                            documento=bet.document,
                            cumpleanos=bet.birthdate,
                            number=bet.number,
                        )),
                    )

                    data = read_message(client_socket)
                    if data is None:
                        break                     

            safe_socket.send_all(client_socket, encode_end())
        except Exception as e:
            logger.error(
                action, logger.LogResult.fail, "messages-amount", message_amount
            )
            raise e

    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                try:
                    self._handle_client(client_socket)
                except Exception:
                    pass
                finally:
                    client_socket.close()

