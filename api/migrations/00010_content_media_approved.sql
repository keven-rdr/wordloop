-- +goose Up
-- Aprova as imagens OpenMoji do seed v0: os arquivos agora estao em web/public/media/openmoji e o hash e o do arquivo.
UPDATE media SET sha256 = '5e9aa542a9e8bce244c326fd2d1427f432ad72e5d8ba92905f3d801d393f6b63', status = 'approved' WHERE storage_key = 'openmoji/1F333.svg' AND status = 'pending';
UPDATE media SET sha256 = 'b4db458333fa917a803c8d36865ae57ae5ea54d2aa0e1f906518d3f0c9bf2bcb', status = 'approved' WHERE storage_key = 'openmoji/1F34E.svg' AND status = 'pending';
UPDATE media SET sha256 = 'd95668fd2f872fa9b8c3adb6b6a1b05376cf412dc99fdd485eb5b56ccf853459', status = 'approved' WHERE storage_key = 'openmoji/1F353.svg' AND status = 'pending';
UPDATE media SET sha256 = 'a1d4819ce6112b37303dabaed0a6cd01101d104f9655058631f7401ebb2098ed', status = 'approved' WHERE storage_key = 'openmoji/1F356.svg' AND status = 'pending';
UPDATE media SET sha256 = '8817d1c5ef6ee76bc9a51fdeade54cb0eb01b6d8c272d05d327c96758c342162', status = 'approved' WHERE storage_key = 'openmoji/1F35A.svg' AND status = 'pending';
UPDATE media SET sha256 = '89181c95686680a25c4769bcd7fc41382dee56371d2c44c57bfe0450a07ef933', status = 'approved' WHERE storage_key = 'openmoji/1F35E.svg' AND status = 'pending';
UPDATE media SET sha256 = '3591c81fda189585742444add6ed17b4377cfeedb87010b447fb61ff48a83585', status = 'approved' WHERE storage_key = 'openmoji/1F373.svg' AND status = 'pending';
UPDATE media SET sha256 = '0efe47c77327eb2c631d5f35f7868b4007a287dec4e0714d26385bc9e3bf83f2', status = 'approved' WHERE storage_key = 'openmoji/1F3C3.svg' AND status = 'pending';
UPDATE media SET sha256 = '6f960c58da749f96f0862bff1f5bc5896dc5e2014227e949625f028af959acbe', status = 'approved' WHERE storage_key = 'openmoji/1F3D9.svg' AND status = 'pending';
UPDATE media SET sha256 = 'edfb5e77a7d324526f76295c0b3eb913006fd0424a0ee3dd0f12b0824225568f', status = 'approved' WHERE storage_key = 'openmoji/1F3E0.svg' AND status = 'pending';
UPDATE media SET sha256 = '0450a151cfbbed8afe1999c0934d5743874ecf578a7e45a4815e81cd6b756820', status = 'approved' WHERE storage_key = 'openmoji/1F3EB.svg' AND status = 'pending';
UPDATE media SET sha256 = '7a7ce0d35b38b61549bf1ab5342df3a443d9974d138fa95b5abdc3ce9b6e8cdf', status = 'approved' WHERE storage_key = 'openmoji/1F41F.svg' AND status = 'pending';
UPDATE media SET sha256 = '950c880495d9668746d3af0cfc533de52a228bce41a9553e68b20016d86ece76', status = 'approved' WHERE storage_key = 'openmoji/1F426.svg' AND status = 'pending';
UPDATE media SET sha256 = 'c0524dc19297fb8cdc521f8f09e2ea777717bb4a85fead116aeab5fd957e9fdc', status = 'approved' WHERE storage_key = 'openmoji/1F431.svg' AND status = 'pending';
UPDATE media SET sha256 = '00cae5ba9efae614fc8ed28aafb11bf8bd10d9a40f57ced09f3a3e2155a554a6', status = 'approved' WHERE storage_key = 'openmoji/1F434.svg' AND status = 'pending';
UPDATE media SET sha256 = 'e29dcecb29e73d442fadbb588a0b2a0d4225452070d3d5154d1ddeaf92434713', status = 'approved' WHERE storage_key = 'openmoji/1F436.svg' AND status = 'pending';
UPDATE media SET sha256 = 'b4ad0a9b5c3308773e0006156a7ab2cdf7e33545cd84c1607860e11d187da7e8', status = 'approved' WHERE storage_key = 'openmoji/1F468.svg' AND status = 'pending';
UPDATE media SET sha256 = 'd84f2fac09bbff1d876ea9b49a1129678ee063c76e78a9ddb6b49e876f36b4f2', status = 'approved' WHERE storage_key = 'openmoji/1F469.svg' AND status = 'pending';
UPDATE media SET sha256 = 'b39a4cd6ba6b8dc064b969cd3a3b6b98eabcc4bd9ec252fcf52559efa5ee5adc', status = 'approved' WHERE storage_key = 'openmoji/1F46A.svg' AND status = 'pending';
UPDATE media SET sha256 = 'ec3895fbd78797679bad1be852d05e659977a010764a7b049822ef7c687ae18c', status = 'approved' WHERE storage_key = 'openmoji/1F4A7.svg' AND status = 'pending';
UPDATE media SET sha256 = 'e78ef793a7acea1241d53017c828db74f14c7bf207a9067dd75021a5c4825e4c', status = 'approved' WHERE storage_key = 'openmoji/1F4B5.svg' AND status = 'pending';
UPDATE media SET sha256 = '12ceb88b0c0b0fd7c5ff65956f106d598e5d44536f7934e7d22937e291665753', status = 'approved' WHERE storage_key = 'openmoji/1F4D6.svg' AND status = 'pending';
UPDATE media SET sha256 = 'a69cdc48cf92ad4d52a58847393a2403c469673d2738fd88de01273817a6f9ab', status = 'approved' WHERE storage_key = 'openmoji/1F4F1.svg' AND status = 'pending';
UPDATE media SET sha256 = '8cc1ef3952c6405b178e191a74e4498312acef6cc56455fabf087471f0bc9812', status = 'approved' WHERE storage_key = 'openmoji/1F600.svg' AND status = 'pending';
UPDATE media SET sha256 = '7539906580ea9bf17ba129522842262dd8ad5c2d367d517eb4f96544a3c02320', status = 'approved' WHERE storage_key = 'openmoji/1F622.svg' AND status = 'pending';
UPDATE media SET sha256 = 'f2bc314639afba633ef21e802c022fa64b6a3fe0e3d3739e5826a816f1895286', status = 'approved' WHERE storage_key = 'openmoji/1F697.svg' AND status = 'pending';
UPDATE media SET sha256 = '01cf0a7caa1c41aa37d8d33c44e7a6dad0eb07b8a03b781f84debfbe9ae218d3', status = 'approved' WHERE storage_key = 'openmoji/1F6AA.svg' AND status = 'pending';
UPDATE media SET sha256 = '6481c97dbd993b85d8ee1dc47ceab345b4bce7017dbd9d4545d6adaad1fec583', status = 'approved' WHERE storage_key = 'openmoji/1F6CF.svg' AND status = 'pending';
UPDATE media SET sha256 = 'a1a2840383bdf9cf3732903e66733ef7f9e5cd6da8212d15d92bfbfca5d1d5af', status = 'approved' WHERE storage_key = 'openmoji/1F95A.svg' AND status = 'pending';
UPDATE media SET sha256 = 'd4d2d7d085dea7979510c84038303861c4480a930cbe1b52249315d3bc89e545', status = 'approved' WHERE storage_key = 'openmoji/1F95B.svg' AND status = 'pending';
UPDATE media SET sha256 = 'ec95448b868077b6427e1f09cfe6df0f2af20139226d2cc65cab3c056d39174f', status = 'approved' WHERE storage_key = 'openmoji/1F975.svg' AND status = 'pending';
UPDATE media SET sha256 = 'c8cc75ffc29471e61f8d4188c0f3d99ffc1764e320c1938e2e08c29eed015aeb', status = 'approved' WHERE storage_key = 'openmoji/1F976.svg' AND status = 'pending';
UPDATE media SET sha256 = '43167ec802840b96f1c0f79edc4ca5bff43b6039c4de037883d4d29bcb1fe334', status = 'approved' WHERE storage_key = 'openmoji/1F9D2.svg' AND status = 'pending';
UPDATE media SET sha256 = '36392b6153d538e37c0022c22da10ab218754d1839955e88a107b766be6fe734', status = 'approved' WHERE storage_key = 'openmoji/23F0.svg' AND status = 'pending';
UPDATE media SET sha256 = 'cf039475b1c737bd355572478013138e602814dae141928f7d50d819080ac4ab', status = 'approved' WHERE storage_key = 'openmoji/2600.svg' AND status = 'pending';
UPDATE media SET sha256 = '7de24e2f229c7b43848225ecdbb8b16c829e6a525e8b84fc6c3e584821027e1d', status = 'approved' WHERE storage_key = 'openmoji/2615.svg' AND status = 'pending';

-- +goose Down
UPDATE media SET sha256 = 'pending:calcular-na-ingestao', status = 'pending' WHERE storage_key LIKE 'openmoji/%';
