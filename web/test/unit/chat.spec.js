describe("Asking for a new room", function(){
	var eros = new starbow.Eros
	var chat = eros.chat

	var mockRoomKey = 'mockRoomKey'
	var room = chat.room(mockRoomKey)

	it("creates a room", function() {
		expect(room).not.toBe(null);
		
	})

	it("with a key", function(){
		expect(room.key).toEqual(mockRoomKey.toLowerCase());
	})

	it("with a name", function(){
		expect(room.name).toEqual(mockRoomKey);
	})
})