#!/usr/bin/env bash
set -xeuo pipefail

# Show progress on pipes if `pv` is installed
# Otherwise use plain cat
if ! command -v pv &> /dev/null
then
    PV="cat"
else
    PV="pv"
fi

mount_partition () {
    local partition_file=$1
    local mountpoint=$2

    # second, try to mount as current user
    if mount -o loop "${partition_file}" "${mountpoint}"; then
        return
    fi

    # third, try to mount with sudo
    sudo mount -o loop "${partition_file}" "${mountpoint}"
    # temporarily change ownership of partition files
    sudo chown -R "${USER}:${USER}" "${mountpoint}"
}

umount_partition () {
    sync
    local mountpoint=$1

    # second, try to umount as current user
    if umount "${mountpoint}"; then
        return
    fi

    # third, try to umount with sudo
    # repair ownership of partition files
    sudo chown -R root:root "${mountpoint}"
    sudo umount "${mountpoint}"
}

# Unpacks finished cloud provider image to recalculate dm-verity hash
unpack () {
    local cloudprovider=$1
    local packed_image=$2
    local unpacked_image=$3

    case $cloudprovider in

    gcp)
        echo "📤 Unpacking GCP image..."
        "${PV}" "$packed_image" | tar -xzf - -O > "$unpacked_image"
        echo "  Unpacked image stored in ${unpacked_image}"
        ;;

    azure)
        echo "📤 Unpacking Azure image..."
        qemu-img convert -p -f vpc -O raw "$packed_image" "$unpacked_image"
        echo "  Unpacked image stored in ${unpacked_image}"
        ;;

    *)
        echo  "unknown cloud provider"
        exit 1
        ;;
    esac
}

get_part_offset () {
    local unpacked_image=$1
    local part_number=$2
    local offset
    offset=$(parted -s "${unpacked_image}" unit s print | sed 's/^ //g' | grep "^${part_number}" | tr -s ' ' | cut -d ' ' -f2)
    local offset=${offset::-1}
    echo "${offset}"
}

get_part_size () {
    local unpacked_image=$1
    local part_number=$2
    local size
    size=$(parted -s "${unpacked_image}" unit s print | sed 's/^ //g' | grep "^${part_number}" | tr -s ' ' | cut -d ' ' -f4)
    local size=${size::-1}
    echo "${size}"
}

extract_partition () {
    local unpacked_image=$1
    local part_number=$2
    local extracted_partition_path=$3

    local part_offset
    part_offset=$(get_part_offset "${unpacked_image}" "${part_number}")
    local part_size
    part_size=$(get_part_size "${unpacked_image}" "${part_number}")
    dd status=progress "if=${unpacked_image}" "of=${extracted_partition_path}" bs=512 "skip=${part_offset}" "count=${part_size}" 2>/dev/null
}

overwrite_partition () {
    local unpacked_image=$1
    local part_number=$2
    local extracted_partition_path=$3

    local part_offset
    part_offset=$(get_part_offset "${unpacked_image}" "${part_number}")
    local part_size
    part_size=$(get_part_size "${unpacked_image}" "${part_number}")
    dd status=progress conv=notrunc "if=${extracted_partition_path}" "of=${unpacked_image}" bs=512 "seek=${part_offset}" "count=${part_size}" 2>/dev/null
}

update_verity () {
    local tmp_dir=$1
    local raw_image=$2
    local boot_mountpoint=${tmp_dir}/boot.mount
    local boot_partition=${tmp_dir}/part_boot.raw
    local root_partition=${tmp_dir}/part_root.raw
    local hashtree_partition=${tmp_dir}/part_hashtree.raw

    echo "⬅️  Extracting partitions..."
    extract_partition "${raw_image}" 3 "${boot_partition}"
    extract_partition "${raw_image}" 4 "${root_partition}"
    extract_partition "${raw_image}" 5 "${hashtree_partition}"

    # recalculate verity hashtree
    veritysetup_out=$(veritysetup format "${root_partition}" "${hashtree_partition}")
    roothash=$(echo "${veritysetup_out}" | grep 'Root hash:' | sed --expression='s/Root hash:\s*//g')
    echo "🧮 Recalculated dm-verity hashtree with roothash ${roothash}"
    # update bootloader kernel cmdline
    mkdir -p "${boot_mountpoint}"
    mount_partition "${boot_partition}" "${boot_mountpoint}"
    sed -i -r "s/verity.sysroot=[[:xdigit:]]+/verity.sysroot=${roothash}/g" "${boot_mountpoint}/loader.1/entries/ostree-1-fedora-coreos.conf"
    echo "✍️  Updated bootloader kernel cmdline to include new dm-verity roothash: $(grep '^options ' "${boot_mountpoint}"/loader.1/entries/ostree-1-fedora-coreos.conf)"
    umount_partition "${boot_mountpoint}"
    rmdir "${boot_mountpoint}"

    echo "➡️  Overwriting partitions..."
    overwrite_partition "${raw_image}" 3 "${boot_partition}"
    overwrite_partition "${raw_image}" 5 "${hashtree_partition}"
}

repack () {
    local cloudprovider=$1
    local unpacked_image=$2
    local packed_image=$3
    local unpacked_image_dir
    unpacked_image_dir=$(dirname "${unpacked_image}")
    local unpacked_image_filename
    unpacked_image_filename=$(basename "${unpacked_image}")
    local tmp_tar_file
    tmp_tar_file=$(mktemp -t verity.XXXXXX.tar)

    case $cloudprovider in

    gcp)
        echo "📥 Repacking GCP image..."
        tar --owner=0 --group=0 -C "${unpacked_image_dir}" -Sch --format=oldgnu -f "${tmp_tar_file}" "${unpacked_image_filename}"
        "${PV}" "${tmp_tar_file}" | pigz -9c  > "${packed_image}"
        rm "${tmp_tar_file}"
        echo "  Repacked image stored in ${packed_image}"
        ;;

    azure)
        echo "📥 Repacking Azure image..."
        qemu-img convert -p -f raw -O vpc -o force_size,subformat=fixed "${unpacked_image}" "$packed_image"
        echo "  Repacked image stored in ${packed_image}"
        ;;

    *)
        echo  "unknown cloud provider"
        exit 1
        ;;
    esac
}

echo "🔁 Recalculating dm-verity hashtree 🌳"
tmp_dir=$(mktemp -d -t verity-XXXXXXXXXX)
raw_image="${tmp_dir}/disk.raw"
unpack "$1" "$2" "${raw_image}"
update_verity "${tmp_dir}" "${raw_image}"
repack "$1" "${raw_image}" "${2}"
rm -r "${tmp_dir}"
