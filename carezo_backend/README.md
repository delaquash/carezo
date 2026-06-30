Payment Bank: Flutterwave Comet (100137272)

Payment Remark: Online Payment via Flutterwave: Txn Ref: (100004260218212827152709277215)

Payment Mode: ONLINE PAYMENT

receipt Staff: flutterwave_api
Payment Date: 18/02/2026
Auto-Payment date: 18/02/2026

docker exec -it carezo-postgres psql -U carezo_user -d carezo_db -f migration_add_image_columns.sql



import React, { useRef, useState } from 'react'
import {
  View, Text, ScrollView, Pressable, FlatList,
  Dimensions, ActivityIndicator, SafeAreaView,
} from 'react-native'
import { Image } from 'expo-image'              // expo-image, not react-native Image
import { router, useLocalSearchParams } from 'expo-router' // useLocalSearchParams reads URL params
import {
  ChevronLeft, MoreHorizontal, Heart, Star,
  Users, Zap, Gauge, Phone, MessageCircle,
  BadgeCheck, ChevronRight, BatteryCharging, ParkingCircle,
} from 'lucide-react-native'
import { useCarDetails } from '@/hooks/useCars'

// screen width is needed to size each image to fill exactly one "page" of the slider
const { width: SCREEN_WIDTH } = Dimensions.get('window')

export default function CarDetails() {

  // useLocalSearchParams reads the :id from the URL /car/[id]
  // when you call router.push({ pathname: '/car/[id]', params: { id: car.id } })
  // expo-router puts car.id into params.id — this line reads it back out
  const { id } = useLocalSearchParams<{ id: string }>()

  // fetch the car whose id matches the URL param
  const { data: car, isLoading } = useCarDetails(id)

  // which dot is active — updated whenever the user swipes to a new image
  const [activeImage, setActiveImage] = useState(0)

  // local liked state — in production this would call a favourites endpoint
  const [liked, setLiked] = useState(false)

  // ref lets us programmatically scroll the FlatList if needed later
  const flatListRef = useRef<FlatList>(null)

  // ── loading state ─────────────────────────────────────────────
  if (isLoading) {
    return (
      <View className="flex-1 bg-bg items-center justify-center">
        <ActivityIndicator size="large" color="#111827" />
      </View>
    )
  }

  // ── not found state ───────────────────────────────────────────
  if (!car) {
    return (
      <View className="flex-1 bg-bg items-center justify-center">
        <Text className="text-muted">Car not found</Text>
        <Pressable onPress={() => router.back()} className="mt-4">
          <Text className="text-primary font-semibold">Go back</Text>
        </Pressable>
      </View>
    )
  }

  // images[] for the carousel — fall back to single image if images array is empty
  // WHY: older car records might only have the single `image` field, not `images[]`
  const images: string[] = car.images?.length
    ? car.images
    : [car.image]

  // ── feature tiles data ───────────────────────────────────────
  // each tile maps to one cell in the 3-column grid in the screenshot
  const features = car.features ? [
    { Icon: Users,           label: 'Capacity',     value: `${car.features.capacity} Seats` },
    { Icon: Zap,             label: 'Engine Out',   value: car.features.engineOutput },
    { Icon: Gauge,           label: 'Max Speed',    value: car.features.maxSpeed     },
    { Icon: Zap,             label: 'Advance',      value: car.features.advance      },
    { Icon: BatteryCharging, label: 'Single Charge', value: car.features.range       },
    { Icon: ParkingCircle,   label: 'Advance',      value: car.features.parking      },
  ] : []

  return (
    <SafeAreaView className="flex-1 bg-bg">
      <ScrollView showsVerticalScrollIndicator={false}>

        {/* ── IMAGE CAROUSEL ──────────────────────────────────────── */}
        <View>
          <FlatList
            ref={flatListRef}
            data={images}
            horizontal                        // makes it scroll left/right
            pagingEnabled                     // snaps each image to fill the screen — key for the slider effect
            showsHorizontalScrollIndicator={false}
            keyExtractor={(_, i) => String(i)}

            // onMomentumScrollEnd fires when the scroll animation finishes (finger lifted)
            // contentOffset.x ÷ SCREEN_WIDTH gives us which image index we landed on
            onMomentumScrollEnd={(e) => {
              const index = Math.round(
                e.nativeEvent.contentOffset.x / SCREEN_WIDTH
              )
              setActiveImage(index)  // update active dot
            }}

            renderItem={({ item }) => (
              <Image
                source={{ uri: item }}
                style={{ width: SCREEN_WIDTH, height: 280 }} // each image fills one "page"
                contentFit="cover"
              />
            )}
          />

          {/* back button and more options overlaid on top of the image */}
          <View
            className="absolute top-4 left-0 right-0 flex-row justify-between px-4"
          >
            <Pressable
              onPress={() => router.back()}
              className="bg-white/90 rounded-full p-2.5"
            >
              <ChevronLeft size={20} color="#111827" />
            </Pressable>
            <Pressable className="bg-white/90 rounded-full p-2.5">
              <MoreHorizontal size={20} color="#111827" />
            </Pressable>
          </View>

          {/* heart button — top right corner of the image */}
          <Pressable
            onPress={() => setLiked(!liked)}
            className="absolute top-4 right-4 bg-white/90 rounded-full p-2.5"
          >
            <Heart
              size={18}
              color={liked ? '#EF4444' : '#9CA3AF'}
              fill={liked ? '#EF4444' : 'transparent'}
            />
          </Pressable>

          {/* ── DOTS INDICATOR ──────────────────────────────────────
              one dot per image — the active dot is wider (pill shape)
              to show which image is currently visible */}
          <View className="flex-row justify-center gap-1.5 mt-3">
            {images.map((_, i) => (
              <View
                key={i}
                style={{
                  height: 7,
                  borderRadius: 99,
                  // active dot is wider pill; inactive dots are small circles
                  width: i === activeImage ? 22 : 7,
                  backgroundColor: i === activeImage ? '#111827' : '#D1D5DB',
                }}
              />
            ))}
          </View>
        </View>

        <View className="px-5 pt-5">

          {/* ── CAR NAME + RATING ───────────────────────────────── */}
          <View className="flex-row items-start justify-between mb-1">
            <Text className="text-2xl font-bold text-primary flex-1 mr-4">
              {car.name}
            </Text>
            <View className="flex-row items-center gap-1 mt-1">
              <Text className="text-sm font-bold text-primary">
                {car.rating.toFixed(1)}
              </Text>
              <Star size={13} color="#F59E0B" fill="#F59E0B" />
            </View>
          </View>

          {/* review count below rating */}
          {car.reviewCount && (
            <Text className="text-xs text-muted mb-3">
              ({car.reviewCount} Reviews)
            </Text>
          )}

          {/* description paragraph */}
          {car.description && (
            <Text className="text-sm text-muted leading-5 mb-5">
              {car.description}
            </Text>
          )}

          <View className="h-px bg-border mb-5" />

          {/* ── DRIVER ROW ───────────────────────────────────────── */}
          {car.driver && (
            <View className="flex-row items-center justify-between mb-5">
              <View className="flex-row items-center gap-3">
                <Image
                  source={{ uri: car.driver.avatar }}
                  style={{ width: 48, height: 48, borderRadius: 24 }}
                  contentFit="cover"
                />
                <View>
                  <View className="flex-row items-center gap-1">
                    <Text className="text-sm font-semibold text-primary">
                      {car.driver.name}
                    </Text>
                    {car.driver.isVerified && (
                      // blue verified badge next to driver name
                      <BadgeCheck size={15} color="#3B82F6" fill="#3B82F6" />
                    )}
                  </View>
                  <Text className="text-xs text-muted">
                    {car.driver.totalTrips} trips
                  </Text>
                </View>
              </View>

              {/* call and message icon buttons */}
              <View className="flex-row gap-2">
                <Pressable
                  className="bg-surface-2 border border-border rounded-full p-2.5"
                >
                  <Phone size={16} color="#6B7280" />
                </Pressable>
                <Pressable
                  className="bg-surface-2 border border-border rounded-full p-2.5"
                >
                  <MessageCircle size={16} color="#6B7280" />
                </Pressable>
              </View>
            </View>
          )}

          <View className="h-px bg-border mb-5" />

          {/* ── CAR FEATURES GRID ────────────────────────────────── */}
          {features.length > 0 && (
            <>
              <Text className="text-base font-bold text-primary mb-4">
                Car features
              </Text>
              {/*  3-column grid — each tile is ~30% wide with gap between */}
              <View
                style={{ flexDirection: 'row', flexWrap: 'wrap', gap: 10 }}
                className="mb-5"
              >
                {features.map((f, i) => (
                  <View
                    key={i}
                    className="bg-surface-2 border border-border rounded-2xl p-3"
                    style={{ width: '30%' }}    // 3 per row with gaps
                  >
                    <f.Icon size={22} color="#9CA3AF" />
                    <Text className="text-[10px] text-muted mt-1.5">
                      {f.label}
                    </Text>
                    <Text className="text-xs font-semibold text-primary mt-0.5">
                      {f.value}
                    </Text>
                  </View>
                ))}
              </View>
            </>
          )}

          <View className="h-px bg-border mb-5" />

          {/* ── REVIEWS ──────────────────────────────────────────── */}
          {car.reviews && car.reviews.length > 0 && (
            <>
              <View className="flex-row items-center justify-between mb-4">
                <Text className="text-base font-bold text-primary">
                  Review ({car.reviewCount ?? car.reviews.length})
                </Text>
                <Pressable className="flex-row items-center gap-1">
                  <Text className="text-sm text-muted">See All</Text>
                  <ChevronRight size={13} color="#9CA3AF" />
                </Pressable>
              </View>

              {/* two reviews side by side — horizontal scroll for overflow */}
              <ScrollView
                horizontal
                showsHorizontalScrollIndicator={false}
                contentContainerStyle={{ gap: 12, paddingBottom: 4 }}
              >
                {car.reviews.slice(0, 3).map((review) => (
                  <View
                    key={review.id}
                    className="bg-surface-2 border border-border rounded-2xl p-4"
                    style={{ width: 200 }}
                  >
                    <View className="flex-row items-center justify-between mb-2">
                      <View className="flex-row items-center gap-2">
                        <Image
                          source={{ uri: review.avatar }}
                          style={{ width: 32, height: 32, borderRadius: 16 }}
                          contentFit="cover"
                        />
                        <Text className="text-xs font-semibold text-primary">
                          {review.user}
                        </Text>
                      </View>
                      <View className="flex-row items-center gap-0.5">
                        <Text className="text-xs font-bold text-primary">
                          {review.rating.toFixed(1)}
                        </Text>
                        <Star size={10} color="#F59E0B" fill="#F59E0B" />
                      </View>
                    </View>
                    <Text
                      className="text-xs text-muted leading-4"
                      numberOfLines={3}
                    >
                      {review.comment}
                    </Text>
                  </View>
                ))}
              </ScrollView>
            </>
          )}

          <View style={{ height: 110 }} />   {/* spacer so Book Now doesn't cover content */}
        </View>
      </ScrollView>

      {/* ── BOOK NOW FOOTER ────────────────────────────────────── */}
      <View
        className="absolute bottom-0 left-0 right-0 px-5 pt-4 pb-8
                   bg-bg border-t border-border"
      >
        <View className="flex-row items-center justify-between mb-3">
          <View>
            <Text className="text-xs text-muted">Total Price</Text>
            <Text className="text-2xl font-bold text-primary">
              ${car.pricePerDay}
              <Text className="text-sm font-normal text-muted">/Day</Text>
            </Text>
          </View>
        </View>
        <Pressable
          onPress={() => router.push({
            pathname: '/booking/create',
            params: { carId: car.id }
          })}
          className="bg-brand rounded-full py-4 flex-row
                     items-center justify-center gap-2"
        >
          <Text className="text-brand-text font-bold text-base">
            Book Now
          </Text>
          <ChevronRight size={18} color="#FFFFFF" />
        </Pressable>
      </View>
    </SafeAreaView>
  )
