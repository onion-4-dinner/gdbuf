#include "register_types.h"
#include "test/proto/dependency.h"
#include "test/proto/nested/deeply/nested.h"
#include "test/proto/gdbuf_test.h"
#include "global_enums.h"
#include <gdextension_interface.h>
#include <godot_cpp/core/defs.hpp>
#include <godot_cpp/godot.hpp>

using namespace godot;

void initialize_gdextension_types(ModuleInitializationLevel p_level){
  if (p_level != MODULE_INITIALIZATION_LEVEL_SCENE)
    return;
  GDREGISTER_CLASS(gdbuf::gdbufgenEnums);
  GDREGISTER_CLASS(gdbuf::dependency::DependencyMessage);
  GDREGISTER_CLASS(gdbuf::nested::DeeplyNestedMessage);
  GDREGISTER_CLASS(gdbuf::gdbuf_test::BasicTestMessage);
  GDREGISTER_CLASS(gdbuf::gdbuf_test::SpecialFieldTypesMessage);
  GDREGISTER_CLASS(gdbuf::gdbuf_test::OneOfMessage);
  GDREGISTER_CLASS(gdbuf::gdbuf_test::GoogleWellKnownTypesMessage);
  GDREGISTER_CLASS(gdbuf::gdbuf_test::MapMessage);
  GDREGISTER_CLASS(gdbuf::gdbuf_test::OuterNestedMessage);
  GDREGISTER_CLASS(gdbuf::gdbuf_test::OuterNestedMessageInnerNestedMessage);
  GDREGISTER_CLASS(gdbuf::gdbuf_test::RepeatedComplexMessage);
  GDREGISTER_CLASS(gdbuf::gdbuf_test::RecursiveMessage);
  GDREGISTER_CLASS(gdbuf::gdbuf_test::ReservedMessage);
  GDREGISTER_CLASS(gdbuf::gdbuf_test::EverythingMessage);
}

void uninitialize_gdextension_types(ModuleInitializationLevel p_level) {
  if (p_level != MODULE_INITIALIZATION_LEVEL_SCENE)
    return;
}

extern "C" {

GDExtensionBool GDE_EXPORT
gdextension_init(GDExtensionInterfaceGetProcAddress p_get_proc_address,
                 GDExtensionClassLibraryPtr p_library,
                 GDExtensionInitialization *r_initialization) {
  godot::GDExtensionBinding::InitObject init_obj(p_get_proc_address, p_library,
                                                 r_initialization);
  init_obj.register_initializer(initialize_gdextension_types);
  init_obj.register_terminator(uninitialize_gdextension_types);
  init_obj.set_minimum_library_initialization_level(
      MODULE_INITIALIZATION_LEVEL_SCENE);
  return init_obj.init();
}
}
