import queue
import socket
import threading
import logger
import safe_socket
from protocol import read_expected, encode_bet_ack, encode_winner, encode_end, WinnerMessage, BATCH, ACK
from lottery import Bet, Lottery
from coordinator import Coordinator
class Server:
    def __init__(self, server_host: str, server_port: int, agency_quorum_min: int) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.lottery = Lottery("bets.csv")
        self.lottery_lock = threading.Lock()
        self.coordinator = Coordinator(agency_quorum_min, self.lottery, self.lottery_lock)

    def _handle_client(self, client_socket):
        action = "handle-client"
        message_amount = 0
        agency = None
        try:
            logger.info(action, logger.LogResult.in_progress)
            while True:
                is_end, data = read_expected(client_socket, BATCH)
                if is_end:
                    break
                agency = data[0].agency
                bets = [ Bet(bet.agency, bet.nombre, bet.apellido, bet.documento, bet.cumpleanos, bet.number)
                    for bet in data ]

                with self.lottery_lock:
                    self.lottery.store_bets(bets)

                message_amount += len(bets)
                safe_socket.send_all(client_socket, encode_bet_ack(len(bets)))

            client_channel = queue.Queue() 
            self.coordinator.get_channel().put((agency, client_channel))

            winners = client_channel.get()

            if not isinstance(winners, list):
                logger.error(action, logger.LogResult.fail, "winners", winners)
                return

            for winner in winners: 
                if not isinstance(winner, Bet):
                    logger.error(action, logger.LogResult.fail, "winner is not a bet", winner)
                    return

                safe_socket.send_all(
                    client_socket,
                    encode_winner(WinnerMessage(
                        nombre=winner.first_name,
                        apellido=winner.last_name,
                        documento=winner.document,
                        cumpleanos=winner.birthdate,
                        number=winner.number,
                    )),
                )

                is_end, _ = read_expected(client_socket, ACK)
                if is_end:
                    break

            safe_socket.send_all(client_socket, encode_end())
        except Exception as e:
            logger.error(
                action, logger.LogResult.fail, "messages-amount", message_amount
            )
            raise e
        finally:
            client_socket.close()


    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            coordinator_thread = threading.Thread(target=self.coordinator.start, daemon=True)
            coordinator_thread.start()
            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                client_thread = threading.Thread(target=self._handle_client, args=(client_socket,), daemon=True)
                client_thread.start()