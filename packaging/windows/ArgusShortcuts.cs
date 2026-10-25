using System;
using System.Text;
using System.Runtime.InteropServices;
using System.Runtime.InteropServices.ComTypes;

// IShellLinkW is required: legacy scripting shortcut wrappers use ANSI fields.
public static class ArgusShortcuts
{
    [ComImport, Guid("00021401-0000-0000-C000-000000000046")]
    private class ShellLink { }

    [ComImport, Guid("000214F9-0000-0000-C000-000000000046"),
     InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
    private interface IShellLinkW
    {
        void GetPath([Out, MarshalAs(UnmanagedType.LPWStr)] StringBuilder path, int count, IntPtr data, uint flags);
        void GetIDList(out IntPtr id);
        void SetIDList(IntPtr id);
        void GetDescription([Out, MarshalAs(UnmanagedType.LPWStr)] StringBuilder value, int count);
        void SetDescription([MarshalAs(UnmanagedType.LPWStr)] string value);
        void GetWorkingDirectory([Out, MarshalAs(UnmanagedType.LPWStr)] StringBuilder value, int count);
        void SetWorkingDirectory([MarshalAs(UnmanagedType.LPWStr)] string value);
        void GetArguments([Out, MarshalAs(UnmanagedType.LPWStr)] StringBuilder value, int count);
        void SetArguments([MarshalAs(UnmanagedType.LPWStr)] string value);
        void GetHotkey(out short value);
        void SetHotkey(short value);
        void GetShowCmd(out int value);
        void SetShowCmd(int value);
        void GetIconLocation([Out, MarshalAs(UnmanagedType.LPWStr)] StringBuilder value, int count, out int index);
        void SetIconLocation([MarshalAs(UnmanagedType.LPWStr)] string value, int index);
        void SetRelativePath([MarshalAs(UnmanagedType.LPWStr)] string value, uint reserved);
        void Resolve(IntPtr window, uint flags);
        void SetPath([MarshalAs(UnmanagedType.LPWStr)] string value);
    }

    public static void Create(string path, string target, string arguments, string directory, string icon)
    {
        object link = new ShellLink();
        try
        {
            IShellLinkW native = (IShellLinkW)link;
            native.SetPath(target);
            native.SetArguments(arguments);
            native.SetWorkingDirectory(directory);
            native.SetDescription("Argus");
            native.SetIconLocation(icon, 0);
            native.SetShowCmd(1);
            ((IPersistFile)link).Save(path, true);
        }
        finally { Marshal.ReleaseComObject(link); }
    }

    public static string[] Read(string path)
    {
        object link = new ShellLink();
        try
        {
            ((IPersistFile)link).Load(path, 0);
            IShellLinkW native = (IShellLinkW)link;
            StringBuilder target = new StringBuilder(32768);
            StringBuilder arguments = new StringBuilder(32768);
            native.GetPath(target, target.Capacity, IntPtr.Zero, 0);
            native.GetArguments(arguments, arguments.Capacity);
            return new string[] { target.ToString(), arguments.ToString() };
        }
        finally { Marshal.ReleaseComObject(link); }
    }
}
