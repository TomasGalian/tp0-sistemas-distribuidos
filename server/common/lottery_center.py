import socket
import logging
import signal
import multiprocessing

from common.utils import Bet, store_bets, has_won, load_bets

fields = ['agency', 'first_name', 'last_name', 'document', 'birthdate', 'number']
ACK = b'\xFF'
WINNERS_SIGNAL = b'\xF0'  # Define a constant for the signal to deliver winners

class LotteryCenter:
    def __init__(self, port, listen_backlog, max_agencies):
        # Initialize lottery socket
        self._lottery_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._lottery_socket.bind(('', port))
        self._lottery_socket.listen(listen_backlog)

        self._max_agencies = int(max_agencies)                                  # Number of agencys in the program

        self._manager = multiprocessing.Manager()                               # Create a Manager to handle resources in diff process
        self._winners_by_agency = self._manager.dict()                          # Shared dictionary for winners
        self._client_sockets = self._manager.list()                             # Shared list for client sockets

        self._barrier_bet = multiprocessing.Barrier(int(max_agencies) + 1)      # Barrier to sync all bets done
        self._barrier_sorteo = multiprocessing.Barrier(int(max_agencies) + 1)   # Barrier to sync all winners

        self._lock_file = multiprocessing.Lock()    # Lock para evitar race conditions en el archivos

    def run(self):
        """
        Run the Lottery Center
        Loterry Center accept new connections and establishes a
        communication with an agency. After agency with communucation
        finishes, servers starts to accept new connections again
        """
        # Register signal handler
        signal.signal(signal.SIGTERM, self.handler_Sigterm)

        processes = []

        # Recive connections and bets. Expecting to receive bets from all agencies
        for _ in range(0, self._max_agencies): 
            client_sock = self.__accept_new_connection()                    # The main process is in charge to acept connections
            p = multiprocessing.Process(target=self.__handle_client_connection, args=(client_sock, self._lock_file, self._barrier_bet, self._barrier_sorteo), name=f"Process-{_+1}")
            p.start()
            processes.append(p)


        self._barrier_bet.wait()                                        # Once all bets are received, the lottery center will start the lottery 
        self._winners_by_agency.update(self.__get_winners_by_agency())  # Update shared dictionary
        self._barrier_sorteo.wait()                                     # Once the lottery is done, the lottery center will send the winners to the agencies    

        for p in processes:                         # Wait for all processes to finish
            p.join()

    def handler_Sigterm(self, sig, frame):
        logging.debug('action: shutdown | result: in_progress')
        for _client_sock in self._client_sockets:
            if (_client_sock):
                _client_sock.close()

        self._lottery_socket.close()
        logging.info('action: shutdown | result: success')
        exit(0)

    def __get_winners_by_agency(self):
        """
        Return the results of the lotery by agency in a dictionary
        """
        with self._lock_file:
            bets_list = load_bets()
        winners = {}
        for bet in bets_list:
            if has_won(bet):
                if bet.agency not in winners:
                    winners[bet.agency] = []
                winners[bet.agency].append(bet)
        logging.info('action: sorteo | result: success')
        return winners

    def __send_winners(self, client_sock, winners):        
        # Supose that len(winners) is less than 255
        bets_serialized = bytearray()
        bets_serialized.extend(len(winners).to_bytes(1, byteorder='big')) 
        client_sock.sendall(bets_serialized)

    def __receive_bet(self, client_sock, total_bets):
        bets = []
        for _ in range(total_bets):
            bet_data = {}
            try:
                for field in fields:
                    len_field_bytes = client_sock.recv(1)
                    if not len_field_bytes:
                        raise ValueError("Incomplete data received for field length.")

                    len_field = ord(len_field_bytes)
                    buffer = client_sock.recv(len_field)
                    while len(buffer) < len_field:
                        additional_data = client_sock.recv(len_field - len(buffer))
                        if not additional_data:
                            raise ValueError("Incomplete data received for field value.")
                        buffer += additional_data

                    bet_data[field] = buffer.decode('utf-8')

                bets.append(Bet(
                    bet_data['agency'],
                    bet_data['first_name'],
                    bet_data['last_name'],
                    bet_data['document'],
                    bet_data['birthdate'],
                    bet_data['number']
                ))
            except (ValueError, KeyError) as e:
                logging.error(f"Error processing bet data: {e}")

        return bets
    
    def __handle_client_connection(self, client_sock, lock_file, barrier_bet, barrier_sorteo):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        logging.debug(f'[{multiprocessing.current_process().name}] action: start_process | result: success')
        try:
            # Use total_bets field to know how many bets will be received
            total_bets = ord(client_sock.recv(1)) 
            while total_bets >= 1:                                  
                bets = self.__receive_bet(client_sock, total_bets)
                with lock_file:                
                    for bet in bets:
                        store_bets([bet])
                logging.info(f'action: apuesta_recibida | result: success | cantidad: {total_bets}')
                    
                logging.debug(f'[{multiprocessing.current_process().name}] action: send_ack | result: in_progress')
                client_sock.send(ACK)

                # Expect the size of the other batch. If it is 0 is because the client finish
                total_bets = ord(client_sock.recv(1))
            
            agency_id = bets[0].agency
            self._client_sockets.append(client_sock)  # Add client socket to the list

            barrier_bet.wait()
            barrier_sorteo.wait()

            action = client_sock.recv(1)
            if action == WINNERS_SIGNAL:  # Use the constant instead of the magic number
                winner_agency = self._winners_by_agency.get(agency_id, [])
                
                self.__send_winners(client_sock, winner_agency)

        except OSError as e:
            logging.info(f'action: apuesta_recibida | result: fail | cantidad: {action}')
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._lottery_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
