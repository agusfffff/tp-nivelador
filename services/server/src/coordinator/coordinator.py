import queue


class Coordinator:
    def __init__(self, agency_quorum_min: int, lottery, lottery_lock) -> None:
        self.agency_quorum_min = agency_quorum_min
        self.response_channel = queue.Queue()
        self.clients_channels = {} 
        self.lottery = lottery
        self.lottery_lock = lottery_lock

    def get_channel(self):
        return self.response_channel

    def start(self):
        while True:
            agency_id, client_channel = self.response_channel.get()

            self.clients_channels[agency_id] = client_channel

            if len(self.clients_channels) >= self.agency_quorum_min:
                with self.lottery_lock:
                    bets = list(self.lottery.load_bets())

                agency_list = list(self.clients_channels.keys())

                winners_by_agency = {
                    agency_id: []
                    for agency_id in self.clients_channels
                }

                for bet in bets: 
                    if bet.agency_id in agency_list and self.lottery.has_won(bet):
                        winners_by_agency[bet.agency_id].append(bet)

                for agency_id, client_channel in self.clients_channels.items(): 
                    client_channel.put(winners_by_agency[agency_id])

                self.clients_channels.clear()