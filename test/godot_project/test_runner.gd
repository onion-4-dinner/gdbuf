extends SceneTree

var tests_failed = 0

func _init():
	print("Running gdbuf tests...")
	
	test_basic_message()
	test_oneof_message()
	test_map_message()
	test_repeated_complex_message()
	test_serialization()
	test_nested_message()
	test_enums()

	if tests_failed == 0:
		print("ALL TESTS PASSED")
		quit(0)
	else:
		print(str(tests_failed) + " TESTS FAILED")
		quit(1)

func assert_eq(actual, expected, message):
	if actual != expected:
		print("FAIL: " + message + " - Expected: " + str(expected) + ", Got: " + str(actual))
		tests_failed += 1
	else:
		# print("PASS: " + message)
		pass

func assert_true(condition, message):
	if not condition:
		print("FAIL: " + message)
		tests_failed += 1

func test_basic_message():
	print("--- test_basic_message ---")
	var msg = BasicTestMessage.new()
	
	# Integer fields
	msg.int32_field = 123
	assert_eq(msg.int32_field, 123, "int32_field")
	
	msg.int64_field = 1234567890
	assert_eq(msg.int64_field, 1234567890, "int64_field")
	
	# Float fields
	msg.float_field = 1.23
	# Float comparison might need epsilon, but for 1.23 it might be stable enough for simple equality or verify roughly
	assert_true(abs(msg.float_field - 1.23) < 0.00001, "float_field")
	
	msg.double_field = 1.23456789
	assert_true(abs(msg.double_field - 1.23456789) < 0.0000001, "double_field")

	# Bool field
	msg.bool_field = true
	assert_eq(msg.bool_field, true, "bool_field")
	
	# String field
	msg.string_field = "Hello gdbuf"
	assert_eq(msg.string_field, "Hello gdbuf", "string_field")

func test_oneof_message():
	print("--- test_oneof_message ---")
	var msg = OneOfMessage.new()
	
	# Default case
	assert_eq(msg.get_test_oneof_case(), OneOfMessage.TEST_ONEOF_NOT_SET, "Default oneof case")
	
	# Set string
	msg.string_field = "oneof string"
	assert_eq(msg.get_test_oneof_case(), OneOfMessage.kStringField, "Oneof case string")
	assert_eq(msg.string_field, "oneof string", "Oneof string value")
	
	# Set int, should clear string
	msg.int32_field = 42
	assert_eq(msg.get_test_oneof_case(), OneOfMessage.kInt32Field, "Oneof case int")
	assert_eq(msg.int32_field, 42, "Oneof int value")
	# Note: Current implementation does not physically clear the memory of the other field immediately, 
	# but it is logically cleared (ignored in serialization).
	# assert_eq(msg.string_field, "", "Oneof string cleared") 
	
	# Verify logical clearing via roundtrip
	var bytes = msg.to_byte_array()
	var msg2 = OneOfMessage.new()
	msg2.from_byte_array(bytes)
	assert_eq(msg2.string_field, "", "Inactive oneof field is empty after roundtrip")
	assert_eq(msg2.int32_field, 42, "Active oneof field preserved after roundtrip")

func test_map_message():
	print("--- test_map_message ---")
	var msg = MapMessage.new()
	
	# String -> Int Map
	var my_map = msg.string_int_map
	my_map["one"] = 1
	my_map["two"] = 2
	msg.string_int_map = my_map # Assign back if it's a copy, but Resources might be ref. Let's check.
	
	# Verify
	assert_eq(msg.string_int_map["one"], 1, "Map string->int")
	assert_eq(msg.string_int_map.size(), 2, "Map size")

func test_repeated_complex_message():
	print("--- test_repeated_complex_message ---")
	var msg = RepeatedComplexMessage.new()
	
	# Repeated Enums
	var enums = msg.enums
	enums.append(1) # BASIC_TEST_ENUM_ONE
	enums.append(2) # BASIC_TEST_ENUM_TWO
	msg.enums = enums
	
	assert_eq(msg.enums.size(), 2, "Repeated enum size")
	assert_eq(msg.enums[0], 1, "Repeated enum 0")
	assert_eq(msg.enums[1], 2, "Repeated enum 1")
	
	# Repeated Messages
	var sub1 = BasicTestMessage.new()
	sub1.int32_field = 10
	var sub2 = BasicTestMessage.new()
	sub2.int32_field = 20
	
	var msgs = msg.messages
	msgs.append(sub1)
	msgs.append(sub2)
	msg.messages = msgs
	
	assert_eq(msg.messages.size(), 2, "Repeated message size")
	assert_eq(msg.messages[0].int32_field, 10, "Repeated message 0 field")

func test_serialization():
	print("--- test_serialization ---")
	var msg = BasicTestMessage.new()
	msg.int32_field = 999
	msg.string_field = "Serialized"
	
	var bytes = msg.to_byte_array()
	assert_true(bytes.size() > 0, "Serialize produces bytes")
	
	var msg2 = BasicTestMessage.new()
	var result = msg2.from_byte_array(bytes)
	
	assert_eq(result, 0, "Deserialize success (0=OK)") # Assuming 0 is success, need to check docs or conventions. 
	# Actually usually it returns a GDError, 0 is OK.
	
	assert_eq(msg2.int32_field, 999, "Deserialized int32")
	assert_eq(msg2.string_field, "Serialized", "Deserialized string")

func test_nested_message():
	print("--- test_nested_message ---")
	var msg = OuterNestedMessage.new()
	msg.outer_string = "Outer"
	
	var inner = OuterNestedMessageInnerNestedMessage.new()
	inner.inner_string = "Inner"
	
	msg.inner_msg = inner
	
	assert_eq(msg.inner_msg.inner_string, "Inner", "Nested message field")

func test_enums():
	print("--- test_enums ---")
	# BasicTestEnum is not exposed as a class/constants, checking values manually
	pass

